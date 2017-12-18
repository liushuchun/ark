/*!
 *  Copyright (c) 2015 by Contributors
 * \file ndarray.cc
 * \brief ndarry module of mxnet
 */
#include <dmlc/io.h>
#include <dmlc/logging.h>
#include <dmlc/registry.h>
#include <mxnet/base.h>
#include <mxnet/ndarray.h>
#include <mxnet/resource.h>
#include <mshadow/tensor.h>
#include "./ndarray_function.h"

namespace dmlc {
DMLC_REGISTRY_ENABLE(::mxnet::NDArrayFunctionReg);
}  // namespace dmlc

namespace mxnet {
/*!
 * \brief run a binary operation
 * \param lhs left operand
 * \param rhs right operand
 * \param out the output ndarray
 * \param binary_op the real
 */
template<typename OP>
void BinaryOp(const NDArray &lhs,
              const NDArray &rhs,
              NDArray *out) {
  // no check if both of them are on cpu
  if (lhs.ctx().dev_mask() != cpu::kDevMask || rhs.ctx().dev_mask() != cpu::kDevMask) {
    CHECK(lhs.ctx() == rhs.ctx()) << "operands context mismatch";
  }
  // if out is none, allocate space
  if (out->is_none()) {
    *out = NDArray(OP::GetShape(lhs.shape(), rhs.shape()), lhs.ctx(), true);
  } else {
    // no check if both of them are on cpu
    if (lhs.ctx().dev_mask() != cpu::kDevMask ||
        out->ctx().dev_mask() != cpu::kDevMask) {
      CHECK(out->ctx() == lhs.ctx()) << "target context mismatch";
    }
    CHECK(out->shape() == OP::GetShape(lhs.shape(), rhs.shape()))
        << "target shape mismatch";
  }
  // important: callback must always capture by value
  NDArray ret = *out;
  // get the const variables
  std::vector<Engine::VarHandle> const_vars;
  if (lhs.var() != ret.var()) const_vars.push_back(lhs.var());
  if (rhs.var() != ret.var()) const_vars.push_back(rhs.var());

  // redirect everything to mshadow operations
  switch (lhs.ctx().dev_mask()) {
    case cpu::kDevMask: {
      Engine::Get()->PushSync([lhs, rhs, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Eval<cpu, OP>(lhs.data(), rhs.data(), &tmp, ctx);
        }, lhs.ctx(), const_vars, {ret.var()});
      break;
    }
#if MXNET_USE_CUDA
    case gpu::kDevMask: {
      Engine::Get()->PushSync([lhs, rhs, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Eval<gpu, OP>(lhs.data(), rhs.data(), &tmp, ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, lhs.ctx(), const_vars, {ret.var()});
      break;
    }
#endif
    default: LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
  }
}

void SetValueOp(const real_t &rhs, NDArray *out) {
  CHECK_NE(out->is_none(), true) << "Set value target must not be empty";
  // important: callback must always capture by value
  NDArray ret = *out;
  switch (ret.ctx().dev_mask()) {
    case cpu::kDevMask: {
      Engine::Get()->PushSync([rhs, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Eval<cpu>(rhs, &tmp, ctx);
        }, ret.ctx(), {}, {ret.var()});
      break;
    }
#if MXNET_USE_CUDA
    case gpu::kDevMask: {
      Engine::Get()->PushSync([rhs, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Eval<gpu>(rhs, &tmp, ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, ret.ctx(), {}, {ret.var()});
      break;
    }
#endif
    default: LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
  }
}

/*!
 * \brief run a binary operation
 * \param lhs left operand
 * \param rhs right operand
 * \param out the output ndarray
 * \param binary_op the real
 */
template<typename OP, bool reverse>
void ScalarOp(const NDArray &lhs,
              const real_t &rhs,
              NDArray *out) {
  if (out->is_none()) {
    *out = NDArray(lhs.shape(), lhs.ctx(), true);
  } else {
    CHECK(out->ctx() == lhs.ctx()) << "target context mismatch";
    CHECK(out->shape() == lhs.shape()) << "target shape mismatch";
  }
  // important: callback must always capture by value
  NDArray ret = *out;
  // get the const variables
  std::vector<Engine::VarHandle> const_vars;
  if (lhs.var() != ret.var()) const_vars.push_back(lhs.var());

  // redirect everything to mshadow operations
  switch (lhs.ctx().dev_mask()) {
    case cpu::kDevMask: {
      Engine::Get()->PushSync([lhs, rhs, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Eval<cpu, OP, reverse>(lhs.data(), rhs, &tmp, ctx);
        }, lhs.ctx(), const_vars, {ret.var()});
      break;
    }
#if MXNET_USE_CUDA
    case gpu::kDevMask: {
      Engine::Get()->PushSync([lhs, rhs, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Eval<gpu, OP, reverse>(lhs.data(), rhs, &tmp, ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, lhs.ctx(), const_vars, {ret.var()});
      break;
    }
#endif
    default: LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
  }
}

void CopyFromTo(const NDArray &from, NDArray *to, int priority) {
  CHECK(from.shape() == to->shape())
      << "operands shape mismatch";
  CHECK(from.shape().ndim() != 0)
      << "source operands have zero dimension shape";
  // important: callback must always capture by value
  NDArray ret = *to;
  int a = from.ctx().dev_mask();
  int b = to->ctx().dev_mask();

  std::vector<Engine::VarHandle> const_vars;
  if (from.var() != ret.var()) const_vars.push_back(from.var());

  if (a == cpu::kDevMask && b == cpu::kDevMask) {
    Engine::Get()->PushSync([from, ret](RunContext ctx) {
        ret.CheckAndAlloc();
        TBlob tmp = ret.data();
        ndarray::Copy<cpu, cpu>(from.data(), &tmp,
                                from.ctx(), ret.ctx(), ctx);
      }, from.ctx(), const_vars, {ret.var()},
      FnProperty::kNormal, priority);
  } else {
#if MXNET_USE_CUDA
    if (a == cpu::kDevMask && b == gpu::kDevMask) {
      Engine::Get()->PushSync([from, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Copy<cpu, gpu>(from.data(), &tmp,
                                  from.ctx(), ret.ctx(), ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, ret.ctx(), const_vars, {ret.var()},
        FnProperty::kCopyToGPU, priority);
    } else if (a == gpu::kDevMask && b == cpu::kDevMask) {
      Engine::Get()->PushSync([from, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Copy<gpu, cpu>(from.data(), &tmp,
                                  from.ctx(), ret.ctx(), ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, from.ctx(), const_vars, {ret.var()},
        FnProperty::kCopyFromGPU, priority);
    } else if (a == gpu::kDevMask && b == gpu::kDevMask) {
      Engine::Get()->PushSync([from, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::Copy<gpu, gpu>(from.data(), &tmp,
                                  from.ctx(), ret.ctx(), ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, from.ctx(), const_vars, {ret.var()},
        FnProperty::kCopyFromGPU, priority);
    } else {
      LOG(FATAL) << "unknown device mask";
    }
#else
    LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
#endif
  }
}

void ElementwiseSum(const std::vector<NDArray> &source, NDArray *out, int priority) {
  std::vector<Engine::VarHandle> const_vars;
  const_vars.reserve(source.size());
  for (size_t i = 0; i < source.size(); ++i) {
    if (source[i].var() != out->var()) {
      const_vars.push_back(source[i].var());
    }
    CHECK_EQ(source[i].shape() , out->shape())
        << "operands shape mismatch";
    if (out->ctx().dev_mask() == cpu::kDevMask) {
      CHECK_EQ(source[i].ctx().dev_mask(),  cpu::kDevMask)
          << "operands context mismatch";
    } else {
      CHECK(source[i].ctx() == out->ctx())
          << "operands context mismatch";
    }
  }
  // important: callback must always capture by value
  NDArray ret = *out;

  switch (out->ctx().dev_mask()) {
    case cpu::kDevMask: {
      Engine::Get()->PushSync([source, ret](RunContext ctx) {
          std::vector<TBlob> source_tblob(source.size());
          for (size_t i = 0; i < source.size(); ++i) {
            source_tblob[i] = source[i].data();
          }
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::ElementwiseSum<cpu>(source_tblob, &tmp, ctx);
        }, out->ctx(), const_vars, {ret.var()},
        FnProperty::kNormal, priority);
      break;
    }
#if MXNET_USE_CUDA
    case gpu::kDevMask: {
      Engine::Get()->PushSync([source, ret](RunContext ctx) {
          std::vector<TBlob> source_tblob(source.size());
          for (size_t i = 0; i < source.size(); ++i) {
            source_tblob[i] = source[i].data();
          }
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::ElementwiseSum<gpu>(source_tblob, &tmp, ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, out->ctx(), const_vars, {ret.var()},
        FnProperty::kNormal, priority);
      break;
    }
#endif
    default: LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
  }
}

void ClipOp(const NDArray &src,
            const real_t &a_min, const real_t &a_max,
            NDArray *out) {
  if (out->is_none()) {
    *out = NDArray(src.shape(), src.ctx(), true);
  } else {
    CHECK(out->ctx() == src.ctx()) << "target context mismatch";
    CHECK(out->shape() == src.shape()) << "target shape mismatch";
  }
  NDArray ret = *out;
  std::vector<Engine::VarHandle> const_vars;
  if (src.var() != ret.var()) const_vars.push_back(src.var());
  switch (src.ctx().dev_mask()) {
    case cpu::kDevMask: {
      Engine::Get()->PushSync([src, a_min, a_max, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::EvalClip<cpu>(src.data(), a_min, a_max, &tmp, ctx);
        }, src.ctx(), const_vars, {ret.var()});
      break;
    }
    #if MXNET_USE_CUDA
    case gpu::kDevMask: {
      Engine::Get()->PushSync([src, a_min, a_max, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::EvalClip<gpu>(src.data(), a_min, a_max, &tmp, ctx);
        }, src.ctx(), const_vars, {ret.var()});
      break;
    }
    #endif
    default: LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
  }
}

inline void CopyFromToSimple(const NDArray &from, NDArray *to) {
  CopyFromTo(from, to, 0);
}

template<typename Distribution>
void SampleOP(const real_t &a,
              const real_t &b,
              NDArray *out) {
  CHECK(!out->is_none());
  Resource resource = ResourceManager::Get()->Request(
      out->ctx(), ResourceRequest::kRandom);
  // important: callback must always capture by value
  NDArray ret = *out;
  // redirect everything to mshadow operations
  switch (out->ctx().dev_mask()) {
    case cpu::kDevMask: {
      Engine::Get()->PushSync([a, b, resource, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::EvalRandom<cpu, Distribution>(a, b, resource, &tmp, ctx);
        }, out->ctx(), {}, {ret.var(), resource.var});
      break;
    }
#if MXNET_USE_CUDA
    case gpu::kDevMask: {
      Engine::Get()->PushSync([a, b, resource, ret](RunContext ctx) {
          ret.CheckAndAlloc();
          TBlob tmp = ret.data();
          ndarray::EvalRandom<gpu, Distribution>(a, b, resource, &tmp, ctx);
          // Wait GPU kernel to complete
          ctx.get_stream<gpu>()->Wait();
        }, out->ctx(), {}, {ret.var(), resource.var});
      break;
    }
#endif
    default: LOG(FATAL) << MXNET_GPU_NOT_ENABLED_ERROR;
  }
}

void SampleUniform(real_t begin, real_t end, NDArray *out) {
  SampleOP<ndarray::UniformDistribution>(begin, end, out);
}

void SampleGaussian(real_t mu, real_t sigma, NDArray *out) {
  SampleOP<ndarray::GaussianDistribution>(mu, sigma, out);
}

void RandomSeed(uint32_t seed) {
  ResourceManager::Get()->SeedRandom(seed);
}

template<typename OP>
inline NDArray BinaryOpRet(const NDArray &lhs,
                           const NDArray &rhs) {
  NDArray ret;
  BinaryOp<OP>(lhs, rhs, &ret);
  return ret;
}

template<typename OP, bool reverse>
inline NDArray ScalarOpRet(const NDArray &lhs,
                           const real_t &rhs) {
  NDArray ret;
  ScalarOp<OP, reverse>(lhs, rhs, &ret);
  return ret;
}

template<typename OP>
inline NDArray &BinaryOpApply(NDArray *dst,
                              const NDArray &src) {
  BinaryOp<OP>(*dst, src, dst);
  return *dst;
}

template<typename OP>
inline NDArray &ScalarOpApply(NDArray *dst,
                             const real_t &src) {
  ScalarOp<OP, false>(*dst, src, dst);
  return *dst;
}

// Binary
NDArray operator+(const NDArray &lhs, const NDArray &rhs) {
  return BinaryOpRet<ndarray::Plus>(lhs, rhs);
}
NDArray operator-(const NDArray &lhs, const NDArray &rhs) {
  return BinaryOpRet<ndarray::Minus>(lhs, rhs);
}
NDArray operator*(const NDArray &lhs, const NDArray &rhs) {
  return BinaryOpRet<ndarray::Mul>(lhs, rhs);
}
NDArray operator/(const NDArray &lhs, const NDArray &rhs) {
  return BinaryOpRet<ndarray::Div>(lhs, rhs);
}
// Scalar
NDArray operator+(const NDArray &lhs, const real_t &rhs) {
  return ScalarOpRet<ndarray::Plus, false>(lhs, rhs);
}
NDArray operator-(const NDArray &lhs, const real_t &rhs) {
  return ScalarOpRet<ndarray::Minus, false>(lhs, rhs);
}
NDArray operator*(const NDArray &lhs, const real_t &rhs) {
  return ScalarOpRet<ndarray::Mul, false>(lhs, rhs);
}
NDArray operator/(const NDArray &lhs, const real_t &rhs) {
  return ScalarOpRet<ndarray::Div, false>(lhs, rhs);
}

// Binary
NDArray &NDArray::operator=(real_t scalar) {
  SetValueOp(scalar, this);
  return *this;
}

NDArray &NDArray::operator+=(const NDArray &src) {
  return BinaryOpApply<ndarray::Plus>(this, src);
}
NDArray &NDArray::operator-=(const NDArray &src) {
  return BinaryOpApply<ndarray::Minus>(this, src);
}
NDArray &NDArray::operator*=(const NDArray &src) {
  return BinaryOpApply<ndarray::Mul>(this, src);
}
NDArray &NDArray::operator/=(const NDArray &src) {
  return BinaryOpApply<ndarray::Div>(this, src);
}
// Scalar
NDArray &NDArray::operator+=(const real_t &src) {
  return ScalarOpApply<ndarray::Plus>(this, src);
}
NDArray &NDArray::operator-=(const real_t &src) {
  return ScalarOpApply<ndarray::Minus>(this, src);
}
NDArray &NDArray::operator*=(const real_t &src) {
  return ScalarOpApply<ndarray::Mul>(this, src);
}
NDArray &NDArray::operator/=(const real_t &src) {
  return ScalarOpApply<ndarray::Div>(this, src);
}

void NDArray::Save(dmlc::Stream *strm) const {
  // save shape
  shape_.Save(strm);
  if (is_none()) return;
  // save context
  Context ctx = this->ctx();
  ctx.Save(strm);
  TBlob save_data;
  NDArray temp;
  if (ctx.dev_mask() != cpu::kDevMask) {
    temp = this->Copy(Context::CPU());
    temp.WaitToRead();
    save_data = temp.data();
  } else {
    this->WaitToRead();
    save_data = this->data();
  }
  // save type flag
  int32_t type_flag = save_data.type_flag_;
  CHECK(type_flag == mshadow::DataType<real_t>::kFlag)
      << "Only support float NDArray so far";
  strm->Write(&type_flag, sizeof(type_flag));
  CHECK(save_data.CheckContiguous());
  // save data: need to change this after more type mask is supported
  size_t type_size = sizeof(real_t);
  strm->Write(save_data.dptr_, type_size * shape_.Size());
}

bool NDArray::Load(dmlc::Stream *strm) {
  // load shape
  TShape shape;
  if (!shape.Load(strm)) return false;
  if (shape.ndim() == 0) {
    *this = NDArray(); return true;
  }
  // load context
  Context ctx;
  if (!ctx.Load(strm)) return false;
  // load type flag
  int32_t type_flag;
  if (strm->Read(&type_flag, sizeof(type_flag)) != sizeof(type_flag)) return false;
  CHECK(type_flag == mshadow::DataType<real_t>::kFlag)
      << "Only support float NDArray so far";
  // load data into CPU
  NDArray temp(shape, Context::CPU());
  TBlob load_data = temp.data();
  size_t type_size = sizeof(real_t);
  size_t nread = type_size * shape.Size();

  if (strm->Read(load_data.dptr_, nread) != nread) return false;
  if (ctx.dev_mask() == cpu::kDevMask) {
    *this = std::move(temp); return true;
  } else {
    *this = temp.Copy(ctx); return true;
  }
}


const uint64_t kMXAPINDArrayListMagic = 0x112;

void NDArray::Save(dmlc::Stream* fo,
                   const std::vector<NDArray>& data,
                   const std::vector<std::string>& names) {
  uint64_t header = kMXAPINDArrayListMagic, reserved = 0;
  fo->Write(&header, sizeof(header));
  fo->Write(&reserved, sizeof(reserved));
  fo->Write(data);
  fo->Write(names);
}

void NDArray::Load(dmlc::Stream* fi,
                   std::vector<NDArray>* data,
                   std::vector<std::string>* keys) {
  uint64_t header, reserved;
  CHECK(fi->Read(&header))
      << "Invalid NDArray file format";
  CHECK(fi->Read(&reserved))
      << "Invalid NDArray file format";
  CHECK(header == kMXAPINDArrayListMagic)
      << "Invalid NDArray file format";
  CHECK(fi->Read(data))
      << "Invalid NDArray file format";
  CHECK(fi->Read(keys))
      << "Invalid NDArray file format";
  CHECK(keys->size() == 0 || keys->size() == data->size())
      << "Invalid NDArray file format";
}

NDArray NDArray::Copy(Context ctx) const {
  NDArray ret(shape(), ctx, true);
  CopyFromTo(*this, &ret);
  return ret;
}

void NDArray::SyncCopyFromCPU(const real_t *data, size_t size) const {
  this->WaitToWrite();
  TShape dshape = this->shape();
  CHECK_EQ(dshape.Size(), size)
      << "Memory size do not match";
  Context ctx = this->ctx();
  TBlob dst = this->data();
  TBlob src((real_t*)data, dshape, cpu::kDevMask); // NOLINT(*)

  RunContext run_ctx;
  run_ctx.stream = nullptr;
  if (ctx.dev_mask() == cpu::kDevMask) {
    ndarray::Copy<cpu, cpu>(src, &dst, Context::CPU(), ctx, run_ctx);
  } else {
#if MXNET_USE_CUDA
    // use empty stream to do sync copy
    // TODO(bing, yutian) consider use a Real Stream, so it is not blocking others
    // Maybe move to engine part
    mshadow::Stream<gpu> zero_stream;
    run_ctx.stream = &zero_stream;
    ndarray::Copy<cpu, gpu>(src, &dst, Context::CPU(), ctx, run_ctx);
#else
    LOG(FATAL) << "GPU is not enabled";
#endif
  }
}

void NDArray::SyncCopyToCPU(real_t *data, size_t size) const {
  this->WaitToRead();
  TShape dshape = this->shape();
  CHECK_EQ(dshape.Size(), size)
      << "Memory size do not match";
  Context ctx = this->ctx();
  TBlob src = this->data();
  TBlob dst(data, dshape, cpu::kDevMask); // NOLINT(*)

  RunContext run_ctx;
  run_ctx.stream = nullptr;
  if (ctx.dev_mask() == cpu::kDevMask) {
    ndarray::Copy<cpu, cpu>(src, &dst, ctx, Context::CPU(), run_ctx);
  } else {
#if MXNET_USE_CUDA
    // use empty stream to do sync copy
    // TODO(bing, yutian) consider use a Real Stream, so it is not blocking others
    // Maybe move to engine part
    mshadow::Stream<gpu> zero_stream;
    run_ctx.stream = &zero_stream;
    ndarray::Copy<gpu, cpu>(src, &dst, ctx, Context::CPU(), run_ctx);
#else
    LOG(FATAL) << "GPU is not enabled";
#endif
  }
}

#if MXNET_PREDICT_ONLY == 0
// register API function
// those with underscore will be registered at NDArray
MXNET_REGISTER_NDARRAY_FUN(_set_value).set_function(SetValueOp);


MXNET_REGISTER_NDARRAY_FUN(_plus).set_function(BinaryOp<ndarray::Plus>);
MXNET_REGISTER_NDARRAY_FUN(_minus).set_function(BinaryOp<ndarray::Minus>);
MXNET_REGISTER_NDARRAY_FUN(_mul).set_function(BinaryOp<ndarray::Mul>);
MXNET_REGISTER_NDARRAY_FUN(_div).set_function(BinaryOp<ndarray::Div>);

MXNET_REGISTER_NDARRAY_FUN(dot).set_function(BinaryOp<ndarray::Dot>)
.describe("Calcuate 2D matrix multiplication");

MXNET_REGISTER_NDARRAY_FUN(_onehot_encode).set_function(BinaryOp<ndarray::OneHotEncode>);

MXNET_REGISTER_NDARRAY_FUN(choose_element_0index)
.set_function(BinaryOp<ndarray::MatChooseRowElem>)
.describe("Choose one element from each line(row for python, column for R/Julia)"
          " in lhs according to index indicated by rhs."
          " This function assume rhs uses 0-based index.");

// register API function
// those with underscore will be registered at NDArray
MXNET_REGISTER_NDARRAY_FUN(_plus_scalar).set_function(ScalarOp<ndarray::Plus, false>);
MXNET_REGISTER_NDARRAY_FUN(_minus_scalar).set_function(ScalarOp<ndarray::Minus, false>);
MXNET_REGISTER_NDARRAY_FUN(_mul_scalar).set_function(ScalarOp<ndarray::Mul, false>);
MXNET_REGISTER_NDARRAY_FUN(_div_scalar).set_function(ScalarOp<ndarray::Div, false>);
// register API function
// scalar, reverse scalar
MXNET_REGISTER_NDARRAY_FUN(_rminus_scalar).set_function(ScalarOp<ndarray::Minus, true>);
MXNET_REGISTER_NDARRAY_FUN(_rdiv_scalar).set_function(ScalarOp<ndarray::Div, true>);

// copy function is special
// that we need to remove kAcceptEmptyMutateTarget from it
MXNET_REGISTER_NDARRAY_FUN(_copyto)
.set_function(CopyFromToSimple)
.set_type_mask(kNDArrayArgBeforeScalar);

// register random number generators
MXNET_REGISTER_NDARRAY_FUN(_random_uniform)
.set_body([](NDArray **u, real_t *s, NDArray **out) {
    SampleUniform(s[0], s[1], out[0]);
  })
.set_num_scalars(2)
.set_num_mutate_vars(1);

MXNET_REGISTER_NDARRAY_FUN(_random_gaussian)
.set_body([](NDArray **u, real_t *s, NDArray **out) {
    SampleGaussian(s[0], s[1], out[0]);
  })
.set_num_scalars(2)
.set_num_mutate_vars(1);

MXNET_REGISTER_NDARRAY_FUN(clip)
.set_type_mask(kNDArrayArgBeforeScalar | kAcceptEmptyMutateTarget)
.set_body([](NDArray **u, real_t *s, NDArray **out) {
    ClipOp(*u[0], s[0], s[1], out[0]);
  })
.set_num_use_vars(1)
.set_num_scalars(2)
.set_num_mutate_vars(1)
.describe("Clip ndarray elements to range (a_min, a_max)")
.add_argument("src", "NDArray", "Source input")
.add_argument("a_min", "real_t", "Minimum value")
.add_argument("a_max", "real_t", "Maximum value");
#endif
}  // namespace mxnet
