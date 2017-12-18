# Convert Caffe Model to Mxnet Format

## Introduction

This is an experimental tool for conversion of Caffe model into mxnet model. There are several limitations to note:
* Please first make sure that there is corresponding operator in mxnet before conversion.
* The tool only supports single input and single output network.
* The tool can only work with the L2LayerParameter in Caffe. For older version, please use the ```upgrade_net_proto_binary``` and ```upgrade_net_proto_text``` in ```tools``` folder of Caffe to upgrate them.

We have verified the results of VGG_16 model and BVLC_googlenet results from Caffe model zoo.

## Notes on Codes
* The core function for converting symbol is in ```convert_symbols.py```. ```proto2script``` converts the prototxt to corresponding python script to generate the symbol. Therefore if you need to modify the auto-generated symbols, you can print out the return value. You can also find the supported layers/operators there.
* The weights are converted in ```convert_model.py```.

## Usage
Run ```python convert_model.py caffe_prototxt caffe_model save_model_name``` to convert the models. Run with ```-h``` for more details of parameters.

