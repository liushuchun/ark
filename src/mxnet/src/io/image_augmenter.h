/*!
 *  Copyright (c) 2015 by Contributors
 * \file image_augmenter_opencv.hpp
 * \brief threaded version of page iterator
 */
#ifndef MXNET_IO_IMAGE_AUGMENTER_H_
#define MXNET_IO_IMAGE_AUGMENTER_H_

#if MXNET_USE_OPENCV
#include <opencv2/opencv.hpp>
#endif
#include <mxnet/base.h>
#include <utility>
#include <string>
#include <algorithm>
#include <vector>
#include "../common/utils.h"

namespace mxnet {
namespace io {
/*! \brief image augmentation parameters*/
struct ImageAugmentParam : public dmlc::Parameter<ImageAugmentParam> {
  /*! \brief whether we do random cropping */
  bool rand_crop;
  /*! \brief whether we do nonrandom croping */
  int crop_y_start;
  /*! \brief whether we do nonrandom croping */
  int crop_x_start;
  /*! \brief [-max_rotate_angle, max_rotate_angle] */
  int max_rotate_angle;
  /*! \brief max aspect ratio */
  float max_aspect_ratio;
  /*! \brief random shear the image [-max_shear_ratio, max_shear_ratio] */
  float max_shear_ratio;
  /*! \brief max crop size */
  int max_crop_size;
  /*! \brief min crop size */
  int min_crop_size;
  /*! \brief max scale ratio */
  float max_random_scale;
  /*! \brief min scale_ratio */
  float min_random_scale;
  /*! \brief min image size */
  float min_img_size;
  /*! \brief max image size */
  float max_img_size;
  /*! \brief rotate angle */
  int rotate;
  /*! \brief filled color while padding */
  int fill_value;
  /*! \brief shape of the image data*/
  TShape data_shape;
  // declare parameters
  DMLC_DECLARE_PARAMETER(ImageAugmentParam) {
    DMLC_DECLARE_FIELD(rand_crop).set_default(false)
        .describe("Augmentation Param: Whether to random crop on the image");
    DMLC_DECLARE_FIELD(crop_y_start).set_default(-1)
        .describe("Augmentation Param: Where to nonrandom crop on y.");
    DMLC_DECLARE_FIELD(crop_x_start).set_default(-1)
        .describe("Augmentation Param: Where to nonrandom crop on x.");
    DMLC_DECLARE_FIELD(max_rotate_angle).set_default(0.0f)
        .describe("Augmentation Param: rotated randomly in [-max_rotate_angle, max_rotate_angle].");
    DMLC_DECLARE_FIELD(max_aspect_ratio).set_default(0.0f)
        .describe("Augmentation Param: denotes the max ratio of random aspect ratio augmentation.");
    DMLC_DECLARE_FIELD(max_shear_ratio).set_default(0.0f)
        .describe("Augmentation Param: denotes the max random shearing ratio.");
    DMLC_DECLARE_FIELD(max_crop_size).set_default(-1)
        .describe("Augmentation Param: Maximum crop size.");
    DMLC_DECLARE_FIELD(min_crop_size).set_default(-1)
        .describe("Augmentation Param: Minimum crop size.");
    DMLC_DECLARE_FIELD(max_random_scale).set_default(1.0f)
        .describe("Augmentation Param: Maxmum scale ratio.");
    DMLC_DECLARE_FIELD(min_random_scale).set_default(1.0f)
        .describe("Augmentation Param: Minimum scale ratio.");
    DMLC_DECLARE_FIELD(max_img_size).set_default(1e10f)
        .describe("Augmentation Param: Maxmum image size after resizing.");
    DMLC_DECLARE_FIELD(min_img_size).set_default(0.0f)
        .describe("Augmentation Param: Minimum image size after resizing.");
    DMLC_DECLARE_FIELD(rotate).set_default(-1.0f)
        .describe("Augmentation Param: Rotate angle.");
    DMLC_DECLARE_FIELD(fill_value).set_default(255)
        .describe("Augmentation Param: Maximum value of illumination variation.");
    DMLC_DECLARE_FIELD(data_shape)
        .set_expect_ndim(3).enforce_nonzero()
        .describe("Dataset Param: Shape of each instance generated by the DataIter.");
  }
};

/*! \brief helper class to do image augmentation */
class ImageAugmenter {
 public:
  // contructor
  ImageAugmenter(void) {
#if MXNET_USE_OPENCV
    rotateM_ = cv::Mat(2, 3, CV_32F);
#endif
  }
  virtual ~ImageAugmenter() {
  }
  virtual void Init(const std::vector<std::pair<std::string, std::string> >& kwargs) {
    std::vector<std::pair<std::string, std::string> > kwargs_left;
    kwargs_left = param_.InitAllowUnknown(kwargs);
    for (size_t i = 0; i < kwargs_left.size(); i++) {
        if (!strcmp(kwargs_left[i].first.c_str(), "rotate_list")) {
          const char* val = kwargs_left[i].second.c_str();
          const char *end = val + strlen(val);
          char buf[128];
          while (val < end) {
            sscanf(val, "%[^,]", buf);
            val += strlen(buf) + 1;
            rotate_list_.push_back(atoi(buf));
          }
        }
    }
  }
#if MXNET_USE_OPENCV
#ifdef _MSC_VER
#define M_PI CV_PI
#endif
  /*!
   * \brief augment src image, store result into dst
   *   this function is not thread safe, and will only be called by one thread
   *   however, it will tries to re-use memory space as much as possible
   * \param src the source image
   * \param source of random number
   * \param dst the pointer to the place where we want to store the result
   */
  virtual cv::Mat Process(const cv::Mat &src,
                          common::RANDOM_ENGINE *prnd) {
    using mshadow::index_t;
    cv::Mat res;

    // normal augmentation by affine transformation.
    if (param_.max_rotate_angle > 0 || param_.max_shear_ratio > 0.0f
        || param_.rotate > 0 || rotate_list_.size() > 0) {
      std::uniform_real_distribution<float> rand_uniform(0, 1);
      // shear
      float s = rand_uniform(*prnd) * param_.max_shear_ratio * 2 - param_.max_shear_ratio;
      // rotate
      int angle = std::uniform_int_distribution<int>(
          -param_.max_rotate_angle, param_.max_rotate_angle)(*prnd);
      if (param_.rotate > 0) angle = param_.rotate;
      if (rotate_list_.size() > 0) {
        angle = rotate_list_[std::uniform_int_distribution<int>(0, rotate_list_.size() - 1)(*prnd)];
      }
      float a = cos(angle / 180.0 * M_PI);
      float b = sin(angle / 180.0 * M_PI);
      // scale
      float scale = rand_uniform(*prnd) *
          (param_.max_random_scale - param_.min_random_scale) + param_.min_random_scale;
      // aspect ratio
      float ratio = rand_uniform(*prnd) *
          param_.max_aspect_ratio * 2 - param_.max_aspect_ratio + 1;
      float hs = 2 * scale / (1 + ratio);
      float ws = ratio * hs;
      // new width and height
      float new_width = std::max(param_.min_img_size,
                                 std::min(param_.max_img_size, scale * src.cols));
      float new_height = std::max(param_.min_img_size,
                                  std::min(param_.max_img_size, scale * src.rows));
      cv::Mat M(2, 3, CV_32F);
      M.at<float>(0, 0) = hs * a - s * b * ws;
      M.at<float>(1, 0) = -b * ws;
      M.at<float>(0, 1) = hs * b + s * a * ws;
      M.at<float>(1, 1) = a * ws;
      float ori_center_width = M.at<float>(0, 0) * src.cols + M.at<float>(0, 1) * src.rows;
      float ori_center_height = M.at<float>(1, 0) * src.cols + M.at<float>(1, 1) * src.rows;
      M.at<float>(0, 2) = (new_width - ori_center_width) / 2;
      M.at<float>(1, 2) = (new_height - ori_center_height) / 2;
      cv::warpAffine(src, temp_, M, cv::Size(new_width, new_height),
                     cv::INTER_LINEAR,
                     cv::BORDER_CONSTANT,
                     cv::Scalar(param_.fill_value, param_.fill_value, param_.fill_value));
      res = temp_;
    } else {
      res = src;
    }

    // crop logic
    if (param_.max_crop_size != -1 || param_.min_crop_size != -1) {
      CHECK(res.cols >= param_.max_crop_size && res.rows >= \
              param_.max_crop_size && param_.max_crop_size >= param_.min_crop_size)
          << "input image size smaller than max_crop_size";
      index_t rand_crop_size =
          std::uniform_int_distribution<index_t>(param_.min_crop_size, param_.max_crop_size)(*prnd);
      index_t y = res.rows - rand_crop_size;
      index_t x = res.cols - rand_crop_size;
      if (param_.rand_crop != 0) {
        y = std::uniform_int_distribution<index_t>(0, y)(*prnd);
        x = std::uniform_int_distribution<index_t>(0, x)(*prnd);
      } else {
        y /= 2; x /= 2;
      }
      cv::Rect roi(x, y, rand_crop_size, rand_crop_size);
      cv::resize(res(roi), res, cv::Size(param_.data_shape[1], param_.data_shape[2]));
    } else {
      CHECK(static_cast<index_t>(res.cols) >= param_.data_shape[1]
            && static_cast<index_t>(res.rows) >= param_.data_shape[2])
          << "input image size smaller than input shape";
      index_t y = res.rows - param_.data_shape[2];
      index_t x = res.cols - param_.data_shape[1];
      if (param_.rand_crop != 0) {
        y = std::uniform_int_distribution<index_t>(0, y)(*prnd);
        x = std::uniform_int_distribution<index_t>(0, x)(*prnd);
      } else {
        y /= 2; x /= 2;
      }
      cv::Rect roi(x, y, param_.data_shape[1], param_.data_shape[2]);
      res = res(roi);
    }
    return res;
  }
#endif

 private:
#if MXNET_USE_OPENCV
  // temporal space
  cv::Mat temp_;
  // rotation param
  cv::Mat rotateM_;
#endif
  // parameters
  ImageAugmentParam param_;
  /*! \brief list of possible rotate angle */
  std::vector<int> rotate_list_;
};
}  // namespace io
}  // namespace mxnet
#endif  // MXNET_IO_IMAGE_AUGMENTER_H_