import caffe
from caffe.proto import caffe_pb2
from google.protobuf import text_format
import argparse
import sys

def readProtoSolverFile(filepath):
    solver_config = caffe.proto.caffe_pb2.NetParameter()
    return readProtoFile(filepath, solver_config)

def readProtoFile(filepath, parser_object):
    file = open(filepath, "r")
    if not file:
        raise self.ProcessException("ERROR (" + filepath + ")!")
    text_format.Merge(str(file.read()), parser_object)
    file.close()
    return parser_object

def proto2script(proto_file):
    proto = readProtoSolverFile(proto_file)
    connection = dict()
    symbols = dict()
    top = dict()
    flatten_count = 0
    symbol_string = ""
    if len(proto.layer):
        layer = proto.layer
    elif len(proto.layers):
        layer = proto.layers
    else:
        raise Exception('Invalid proto file.')

    # We assume the first bottom blob of first layer is the output from data layer
    input_name = layer[0].bottom[0]
    output_name = ""
    mapping = {input_name : 'data'}
    need_flatten = {input_name : False}
    for i in range(len(layer)):
        type_string = ''
        param_string = ''
        name = layer[i].name.replace('/', '_')
        if layer[i].type == 'Convolution' or layer[i].type == 4:
            type_string = 'mx.symbol.Convolution'
            param = layer[i].convolution_param
            pad = 0 if len(param.pad) == 0 else param.pad[0]
            stride = 1 if len(param.stride) == 0 else param.stride[0]
            param_string = "num_filter=%d, pad=(%d,%d), kernel=(%d,%d), stride=(%d,%d), no_bias=%s" %\
                (param.num_output, pad, pad, param.kernel_size[0],\
                param.kernel_size[0], stride, stride, not param.bias_term)
            need_flatten[name] = True
        if layer[i].type == 'Pooling' or layer[i].type == 17:
            type_string = 'mx.symbol.Pooling'
            param = layer[i].pooling_param
            param_string = "pad=(%d,%d), kernel=(%d,%d), stride=(%d,%d)" %\
                (param.pad, param.pad, param.kernel_size,\
                param.kernel_size, param.stride, param.stride)
            if param.pool == 0:
                param_string = param_string + ", pool_type='max'"
            elif param.pool == 1:
                param_string = param_string + ", pool_type='avg'"
            else:
                raise Exception("Unknown Pooling Method!")
            need_flatten[name] = True
        if layer[i].type == 'ReLU' or layer[i].type == 18:
            type_string = 'mx.symbol.Activation'
            param_string = "act_type='relu'"
            need_flatten[name] = need_flatten[mapping[layer[i].bottom[0]]]
        if layer[i].type == 'LRN' or layer[i].type == 15:
            type_string = 'mx.symbol.LRN'
            param = layer[i].lrn_param
            param_string = "alpha=%f, beta=%f, knorm=%f, nsize=%d" %\
                (param.alpha, param.beta, param.k, param.local_size)
            need_flatten[name] = True
        if layer[i].type == 'InnerProduct' or layer[i].type == 14:
            type_string = 'mx.symbol.FullyConnected'
            param = layer[i].inner_product_param
            param_string = "num_hidden=%d, no_bias=%s" % (param.num_output, not param.bias_term)
            need_flatten[name] = False
        if layer[i].type == 'Dropout' or layer[i].type == 6:
            type_string = 'mx.symbol.Dropout'
            param = layer[i].dropout_param
            param_string = "p=%f" % param.dropout_ratio
            need_flatten[name] = need_flatten[mapping[layer[i].bottom[0]]]
        if layer[i].type == 'Softmax' or layer[i].type == 20:
            type_string = 'mx.symbol.SoftmaxOutput'

            # We only support single output network for now.
            output_name = name
        if layer[i].type == 'Flatten' or layer[i].type == 8:
            type_string = 'mx.symbol.Flatten'
            need_flatten[name] = False
        if layer[i].type == 'Split' or layer[i].type == 22:
            type_string = 'split'
        if layer[i].type == 'Concat' or layer[i].type == 3:
            type_string = 'mx.symbol.Concat'
            need_flatten[name] = True
        if type_string == '':
            raise Exception('Unknown Layer %s!' % layer[i].type)

        if type_string != 'split':
            bottom = layer[i].bottom
            if param_string != "":
                param_string = ", " + param_string
            if len(bottom) == 1:
                if need_flatten[mapping[bottom[0]]] and type_string == 'mx.symbol.FullyConnected':
                    flatten_name = "flatten_%d" % flatten_count
                    symbol_string += "%s=mx.symbol.Flatten(name='%s', data=%s)\n" %\
                        (flatten_name, flatten_name, mapping[bottom[0]])
                    flatten_count += 1
                    need_flatten[flatten_name] = False
                    bottom[0] = flatten_name
                    mapping[bottom[0]] = bottom[0]
                symbol_string += "%s = %s(name='%s', data=%s %s)\n" %\
                    (name, type_string, name, mapping[bottom[0]], param_string)
            else:
                symbol_string += "%s = %s(name='%s', *[%s] %s)\n" %\
                    (name, type_string, name, ','.join([mapping[x] for x in bottom]), param_string)
        for j in range(len(layer[i].top)):
            mapping[layer[i].top[j]] = name
    return symbol_string, output_name

def proto2symbol(proto_file):
    sym, output_name = proto2script(proto_file)
    sym = "import mxnet as mx\n" \
            + "data = mx.symbol.Variable(name='data')\n" \
            + sym
    exec(sym)
    exec("ret = " + output_name)
    return ret

def main():
    symbol_string, output_name = proto2script(sys.argv[1])
    if len(sys.argv) > 2:
        with open(sys.argv[2], 'w') as fout:
            fout.write(symbol_string)
    else:
        print(symbol_string)

if __name__ == '__main__':
    main()
