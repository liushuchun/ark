FROM nvidia/cuda:8.0-cudnn6-devel
LABEL maintainer "Qiniu OCR team"

RUN sed -i s/archive.ubuntu.com/mirrors.163.com/g /etc/apt/sources.list
RUN sed -i s/security.ubuntu.com/mirrors.163.com/g /etc/apt/sources.list
# 这两个 NVIDIA source list 更新存在问题
RUN rm /etc/apt/sources.list.d/cuda.list /etc/apt/sources.list.d/nvidia-ml.list

# apt-get && python && pip3
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates wget vim lrzsz curl git unzip build-essential cmake \
    python-dev python3-pip python-tk \
    libatlas-base-dev libopencv-dev libcurl4-openssl-dev \
    libgtest-dev \
    openssh-server rsync && \
    cd /usr/src/gtest && cmake CMakeLists.txt && make && cp *.a /usr/lib && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# pip3
RUN pip3 install --upgrade pip && pip3 --no-cache-dir --default-timeout=6000 install -i http://pypi.douban.com/simple/ --trusted-host=pypi.douban.com -U pip setuptools && \
    pip3 --no-cache-dir --default-timeout=6000 install -i http://pypi.douban.com/simple/ --trusted-host=pypi.douban.com hickle nose pylint pyyaml numpy nose-timer requests easydict matplotlib cython scikit-image

# opencv 3
RUN export OPENCV_CONTRIB_ROOT=/tmp/opencv-contrib OPENCV_ROOT=/tmp/opencv OPENCV_VER=3.2.0 && \
    git clone -b ${OPENCV_VER} --depth 1 https://github.com/opencv/opencv.git ${OPENCV_ROOT} && \
    git clone -b ${OPENCV_VER} --depth 1 https://github.com/opencv/opencv_contrib.git ${OPENCV_CONTRIB_ROOT} && \
    mkdir -p ${OPENCV_ROOT}/build && cd ${OPENCV_ROOT}/build && \
    cmake -D CMAKE_BUILD_TYPE=RELEASE -D CMAKE_INSTALL_PREFIX=/usr/local \
    -D OPENCV_ICV_URL="http://devtools.dl.atlab.ai/docker/" \
    -D OPENCV_PROTOBUF_URL="http://devtools.dl.atlab.ai/docker/" \
    -D OPENCV_CONTRIB_BOOSTDESC_URL="http://devtools.dl.atlab.ai/docker/" \
    -D OPENCV_CONTRIB_VGG_URL="http://devtools.dl.atlab.ai/docker/" \
    -D INSTALL_C_EXAMPLES=OFF -D INSTALL_PYTHON_EXAMPLES=OFF \
    -D OPENCV_EXTRA_MODULES_PATH=${OPENCV_CONTRIB_ROOT}/modules \
    -D WITH_CUDA=ON -D BUILD_opencv_python3=ON -D BUILD_EXAMPLES=OFF .. && \
    make -j"$(nproc)" && make install && ldconfig && \
    rm -rf /tmp/*

# fcis
ENV FCIS_ROOT=/opt/fcis
ENV PYFCIS_ROOT=$FCIS_ROOT/fcis:$FCIS_ROOT/lib
RUN git clone -b master --depth 1 https://github.com/msracver/FCIS.git ${FCIS_ROOT} && \
    cd $FCIS_ROOT && sh init.sh

# warp-ctc
ENV WARP_CTC_ROOT=/opt/warp-ctc
RUN git clone https://github.com/baidu-research/warp-ctc.git ${WARP_CTC_ROOT} && cd ${WARP_CTC_ROOT} && mkdir build && cd build && cmake .. && make -j4 && make install


# mxnet
ENV MXNET_ROOT=/opt/mxnet MXNET_VER=v1.0.0
RUN mkdir -p ${MXNET_ROOT} && cd ${MXNET_ROOT} && git clone -b ${MXNET_VER} --depth 1 --recursive https://github.com/ataraxialab/incubator-mxnet.git . &&\
 echo "WARPCTC_PATH = $(WARP_CTC_ROOT)" >>config.mk && \
 echo "MXNET_PLUGINS += plugin/warpctc/warpctc.mk" >>config.mk && \
 cp ${FCIS_ROOT}/fcis/operator_cxx/* ${MXNET_ROOT}/src/operator/contrib -r && \
 make -j"$(nproc)" USE_DIST_KVSTORE=1   USE_CUDA=1 USE_CUDA_PATH=/usr/local/cuda USE_CUDNN=1 \
 EXTRA_OPERATORS=${MXNET_ROOT}/example/rcnn/operator && \
 rm -rf build && cd ${MXNET_ROOT}/example/rcnn && make

ENV PYTHONPATH $MXNET_ROOT/python:$PYFCIS_ROOT:$PYTHONPATH
ENV PATH /usr/local/cuda/bin:$PATH

# 增加 dumb-init 和 entrypoint.sh 脚本
# file stored in qiniu://avatest@qiniu.com@z0/devtools/docker/dumb-init_1.2.0_amd64
RUN wget -O /usr/local/bin/dumb-init http://devtools.dl.atlab.ai/docker/dumb-init_1.2.0_amd64 && \
    mkdir /ava && wget -O /ava/entrypoint.sh http://devtools.dl.atlab.ai/docker/entrypoint.sh && \
    chmod +x /usr/local/bin/dumb-init /ava/entrypoint.sh
ENTRYPOINT ["/ava/entrypoint.sh"]

# 将时区改成 GMT+8
RUN wget -O /tmp/PRC-tz http://devtools.dl.atlab.ai/docker/PRC-tz && mv /tmp/PRC-tz /etc/localtime
ENV LC_ALL=C.UTF-8
LABEL com.qiniu.atlab.os = "ubuntu-16.04"
LABEL com.qiniu.atlab.type = "mxnet"






