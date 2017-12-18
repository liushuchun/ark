// mexnet.cc

#define MSHADOW_FORCE_STREAM
#ifndef MSHADOW_USE_CBLAS
#define MSHADOW_USE_CBLAS 	1
#endif
#define MSHADOW_USE_CUDA 	0
#define MSHADOW_USE_MKL 	0
#define MSHADOW_RABIT_PS 	0
#define MSHADOW_DIST_PS 	0

#define MXNET_USE_OPENCV 	0
#define MXNET_PREDICT_ONLY 	1
#define DISABLE_OPENMP 1

#include "src/ndarray/unary_function.cc"
#include "src/ndarray/ndarray_function.cc"
#include "src/ndarray/ndarray.cc"
#include "src/engine/engine.cc"
#include "src/engine/naive_engine.cc"
#include "src/symbol/graph_executor.cc"
#include "src/symbol/static_graph.cc"
#include "src/symbol/symbol.cc"
#include "src/operator/operator.cc"
#include "src/operator/activation.cc"
#include "src/operator/batch_norm.cc"
#include "src/operator/block_grad.cc"
#include "src/operator/concat.cc"
#include "src/operator/convolution.cc"
#include "src/operator/dropout.cc"
#include "src/operator/elementwise_binary_op.cc"
#include "src/operator/elementwise_sum.cc"
#include "src/operator/fully_connected.cc"
#include "src/operator/leaky_relu.cc"
#include "src/operator/lrn.cc"
#include "src/operator/pooling.cc"
#include "src/operator/regression_output.cc"
#include "src/operator/reshape.cc"
#include "src/operator/slice_channel.cc"
#include "src/operator/softmax_output.cc"
#include "src/operator/deconvolution.cc"
#include "src/storage/storage.cc"
#include "src/common/tblob_op_registry.cc"

#include "src/resource.cc"

#include "src/c_api/c_predict_api.cc"
#include "src/c_api/c_api_error.cc"
