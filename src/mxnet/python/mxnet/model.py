# pylint: disable=fixme, invalid-name, too-many-arguments, too-many-locals
# pylint: disable=too-many-branches, too-many-statements
"""MXNet model module"""
from __future__ import absolute_import

import numpy as np
import time
import logging
from . import io
from . import nd
from . import symbol as sym
from . import optimizer as opt
from . import metric
from . import kvstore as kvs
from .context import Context, cpu
from .initializer import Uniform
from collections import namedtuple
from .optimizer import get_updater

BASE_ESTIMATOR = object

try:
    from sklearn.base import BaseEstimator
    BASE_ESTIMATOR = BaseEstimator
except ImportError:
    SKLEARN_INSTALLED = False

# Parameter to pass to batch_end_callback
BatchEndParam = namedtuple('BatchEndParams',
                           ['epoch',
                            'nbatch',
                            'eval_metric'])


def _check_arguments(symbol):
    """Check the argument names of symbol.

    This function checks the duplication of arguments in Symbol.
    The check is done for feedforward net for now.

    Parameters
    ----------
    symbol : Symbol
        The network configuration

    Returns
    -------
    data_index : int
        Index position of data.
    label_index : int
        Index position of label
    """
    arg_names = symbol.list_arguments()
    data_index, label_index = None, None
    arg_set = set()
    for index, name in enumerate(arg_names):
        if name.endswith('label'):
            if label_index is not None:
                raise ValueError('Two arguments with suffix \"label\", ' +
                                 'only accept one label in config for now, '+
                                 'arguments are %s' % str(arg_names))
            label_index = index
        if name.endswith('data'):
            if data_index is not None:
                raise ValueError('Two arguments with suffix \"label\", ' +
                                 'only accept one input data in config for now, ' +
                                 'arguments are %s' % str(arg_names))
            data_index = index
        if name in arg_set:
            raise ValueError(('Find duplicated argument name \"%s\", ' +
                              'please make the weight name non-duplicated(using name arguments), ' +
                              'arguments are %s') % (name, str(arg_names)))
        arg_set.add(name)

    aux_names = symbol.list_auxiliary_states()
    for name in aux_names:
        if name in arg_set:
            raise ValueError(
                ('Find duplicated auxiliary param name \"%s\", ' +
                 'please make the weight name non-duplicated(using name arguments), ' +
                 'arguments are %s, auxiliary params are %s'
                ) % (name, str(arg_names), str(aux_names)))

    return (data_index, label_index)


def _split_input_slice(input_shape, num_split):
    """Get input slice from the input shape.

    Parameters
    ----------
    input_shape : tuple
        The input shape of the net.

    num_split : int
        The number of split we want to have.

    Returns
    -------
    slices : list of slice
        The split slices to get a specific slice.

    shapes : list of tuples
        The shape of each split slice.

    Raises
    ------
    ValueError
        If there are two many splits such that some slice can be empty.
    """
    batch_size = input_shape[0]
    step = (batch_size + num_split - 1) / num_split
    slices = []
    shapes = []
    for k in range(num_split):
        begin = int(min(k * step, batch_size))
        end = int(min((k+1) * step, batch_size))
        if begin == end:
            raise ValueError('Too many slices such that some splits are empty')
        slices.append(slice(begin, end))
        s = list(input_shape)
        s[0] = end - begin
        shapes.append(tuple(s))
    return (slices, shapes)

def _create_kvstore(kvstore, num_device, arg_params):
    """Create kvstore

    This function select and create a proper kvstore if given the kvstore type

    Parameters
    ----------

    kvstore : KVStore or str
        The kvstore

    num_device : int
        The number of devices

    arg_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's weights.
    """

    if isinstance(kvstore, kvs.KVStore):
        kv = kvstore
    elif isinstance(kvstore, str):
        # create kvstore using the string type
        if num_device is 1 and 'dist' not in kvstore:
            # no need to use kv for single device and single machine
            kv = None
        else:
            if kvstore is 'local':
                # automatically select a proper local
                max_size = max(np.prod(param.shape) for param in arg_params.values())
                if max_size < 1024 * 1024 * 16:
                    kvstore = 'local_update_cpu'
                else:
                    kvstore = 'local_allreduce_cpu'
                logging.info('Auto-select kvstore type = %s', kvstore)
            kv = kvs.create(kvstore)
    else:
        raise TypeError('kvstore must be either KVStore or str')

    # detect whether or not update weight on kvstore
    update_on_kvstore = True
    if not kv or 'local_allreduce' in kv.type:
        update_on_kvstore = False

    return (kv, update_on_kvstore)

def _train_multi_device(symbol, ctx, input_shape,
                        arg_params, aux_params,
                        begin_epoch, end_epoch, optimizer,
                        kvstore, update_on_kvstore,
                        train_data, eval_data=None, eval_metric=None,
                        epoch_end_callback=None, batch_end_callback=None,
                        logger=None):
    """Internal training function on multiple devices.

    This function will also work for single device as well.

    Parameters
    ----------
    symbol : Symbol
        The network configuration

    ctx : list of Context
        The training devices.

    input_shape : tuple
        Shape of input data batch.

    arg_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's weights.

    aux_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's auxiliary states.

    begin_epoch : int
        The begining training epoch.

    end_epoch : int
        The end training epoch.

    optimizer : Optimizer
        The optimization algorithm

    train_data : DataIter
        Training data iterator.

    eval_data : DataIter
        Validation data iterator.

    eval_metric : EvalMetric
        A evaluation function.

    epoch_end_callback : callable(epoch, symbol, arg_params, aux_states)
        A callback that is invoked at end of each epoch.
        This can be used to checkpoint model each epoch.

    batch_end_callback : callable(BatchEndParams)
        A callback that is invoked at end of each batch.
        This can be used to measure speed, get result from evaluation metric. etc.

    kvstore : KVStore
        The KVStore

    update_on_kvstore : bool
        whether or not perform weight updating on kvstore

    logger : logging logger
        When not specified, default logger will be used.

    Notes
    -----
    - This function will inplace update the NDArrays in arg_parans and aux_states.
    """
    if logger is None:
        logger = logging
    # preparation
    num_device = len(ctx)
    logging.info('Start training with %s', str(ctx))

    slices, shapes = _split_input_slice(input_shape, num_device)
    train_execs = [symbol.simple_bind(ctx=c, data=s, grad_req='write')
                   for c, s in zip(ctx, shapes)]
    arg_names = symbol.list_arguments()
    aux_names = symbol.list_auxiliary_states()
    # data structure
    arg_blocks = [
        [x.arg_arrays[index] for x in train_execs]
        for index in range(len(train_execs[0].arg_arrays))]
    grad_blocks = [
        [x.grad_arrays[index] for x in train_execs]
        for index in range(len(train_execs[0].grad_arrays))]
    aux_blocks = [
        [x.aux_arrays[index] for x in train_execs]
        for index in range(len(train_execs[0].aux_arrays))]

    for texec in train_execs:
        texec.copy_params_from(arg_params, aux_params)

    # init optmizer
    optimizer.begin_epoch(begin_epoch)

    if not update_on_kvstore:
        updater = get_updater(optimizer)

    # init kvstore
    if kvstore:
        # init optimizer
        if update_on_kvstore:
            kvstore.set_optimizer(optimizer)

        # init kv
        for index, pair in enumerate(zip(arg_blocks, grad_blocks)):
            arg_list, grad_list = pair
            if grad_list[0] is not None:
                kvstore.init(index, arg_list[0])

                # pull the weight back
                if update_on_kvstore:
                    kvstore.pull(index, arg_list, priority=-index)

    # Input and output data structure
    data_index, label_index = _check_arguments(symbol)
    merged_shape = list(train_execs[0].outputs[0].shape)
    merged_shape[0] = input_shape[0]
    merged_shape = tuple(merged_shape)
    out_cpu_array = nd.zeros(merged_shape, cpu())

    # Now start training
    for epoch in range(begin_epoch, end_epoch):
        # Training phase
        tic = time.time()
        eval_metric.reset()
        nbatch = 0
        # Iterate over training data.
        for data, label in train_data:
            # Copy data into the target
            for target, islice in zip(arg_blocks[label_index], slices):
                label[islice].copyto(target)
            for target, islice in zip(arg_blocks[data_index], slices):
                data[islice].copyto(target)
            # forward backward pass
            for texec, islice in zip(train_execs, slices):
                texec.forward(is_train=True)
                texec.outputs[0].copyto(out_cpu_array[islice])
            for texec in train_execs:
                texec.backward()
            # update the parameters
            for index, pair in enumerate(zip(arg_blocks, grad_blocks)):
                arg_list, grad_list = pair
                if grad_list[0] is None:
                    continue
                # Gradient synchronization
                if kvstore:
                    # push gradient, priority is negative index
                    kvstore.push(index, grad_list, priority=-index)
                    if update_on_kvstore:
                        # pull back the weights
                        kvstore.pull(index, arg_list, priority=-index)
                    else:
                        # pull back the sum gradients, to the same locations.
                        kvstore.pull(index, grad_list, priority=-index)
                if not update_on_kvstore:
                    for k, p in enumerate(zip(arg_list, grad_list)):
                        # faked an index here, to make optimizer create diff
                        # state for the same index but on diff devs, TODO(mli)
                        # use a better solution latter
                        w, g = p
                        updater(index*num_device+k, g, w)

            nbatch += 1
            # batch callback (for print purpose)
            if batch_end_callback != None:
                batch_end_params = BatchEndParam(epoch=epoch,
                                                 nbatch=nbatch,
                                                 eval_metric=eval_metric)
                if isinstance(batch_end_callback, list):
                    for call in batch_end_callback:
                        call(batch_end_params)
                else:
                    batch_end_callback(batch_end_params)
            # evaluate at end, so out_cpu_array can lazy copy
            eval_metric.update(label, out_cpu_array)

        # reset training data after epoch finish
        train_data.reset()
        name, value = eval_metric.get()
        logger.info('Epoch[%d] Train-%s=%f', epoch, name, value)
        toc = time.time()
        logger.info('Epoch[%d] Time cost=%.3f', epoch, (toc - tic))
        # evaluation
        if eval_data:
            eval_metric.reset()
            for data, label in eval_data:
                # Copy data into the target
                for target, islice in zip(arg_blocks[label_index], slices):
                    label[islice].copyto(target)
                for target, islice in zip(arg_blocks[data_index], slices):
                    data[islice].copyto(target)
                # forward pass
                for texec, islice in zip(train_execs, slices):
                    texec.forward(is_train=False)
                    texec.outputs[0].copyto(out_cpu_array[islice])
                eval_metric.update(label, out_cpu_array)
            eval_data.reset()
            name, value = eval_metric.get()
            logger.info('Epoch[%d] Validation-%s=%f', epoch, name, value)

        if epoch_end_callback or epoch + 1 == end_epoch:
            # copy data back to cpu
            for name, block in zip(arg_names, arg_blocks):
                if name in arg_params:
                    weight = sum(w.copyto(cpu()) for w in block) / len(block)
                    weight.copyto(arg_params[name])
            for name, block in zip(aux_names, aux_blocks):
                if name in aux_params:
                    weight = sum(w.copyto(cpu()) for w in block) / len(block)
                    weight.copyto(aux_params[name])
        if epoch_end_callback != None:
            if isinstance(epoch_end_callback, list):
                for call in epoch_end_callback:
                    call(epoch, symbol, arg_params, aux_params)
            else:
                epoch_end_callback(epoch, symbol, arg_params, aux_params)
    # end of all epochs
    return


def save_checkpoint(prefix, epoch, symbol, arg_params, aux_params):
    """Checkpoint the model data into file.

    Parameters
    ----------
    prefix : str
        Prefix of model name.

    epoch : int
        The epoch number of the model.

    symbol : Symbol
        The input symbol

    arg_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's weights.

    aux_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's auxiliary states.

    Notes
    -----
    - ``prefix-symbol.json`` will be saved for symbol.
    - ``prefix-epoch.params`` will be saved for parameters.
    """
    symbol.save('%s-symbol.json' % prefix)
    save_dict = {('arg:%s' % k) : v for k, v in arg_params.items()}
    save_dict.update({('aux:%s' % k) : v for k, v in aux_params.items()})
    param_name = '%s-%04d.params' % (prefix, epoch)
    nd.save(param_name, save_dict)
    logging.info('Saved checkpoint to \"%s\"', param_name)


def load_checkpoint(prefix, epoch):
    """Load model checkpoint from file.

    Parameters
    ----------
    prefix : str
        Prefix of model name.

    epoch : int
        Epoch number of model we would like to load.

    Returns
    -------
    symbol : Symbol
        The symbol configuration of computation network.

    arg_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's weights.

    aux_params : dict of str to NDArray
        Model parameter, dict of name to NDArray of net's auxiliary states.

    Notes
    -----
    - ``prefix-symbol.json`` will be saved for symbol.
    - ``prefix-epoch.params`` will be saved for parameters.
    """
    symbol = sym.load('%s-symbol.json' % prefix)
    save_dict = nd.load('%s-%04d.params' % (prefix, epoch))
    arg_params = {}
    aux_params = {}
    for k, v in save_dict.items():
        tp, name = k.split(':', 1)
        if tp == 'arg':
            arg_params[name] = v
        if tp == 'aux':
            aux_params[name] = v
    return (symbol, arg_params, aux_params)


class FeedForward(BASE_ESTIMATOR):
    """Model class of MXNet for training and predicting feedforward nets.

    This class is designed for a single-data single output supervised network.

    Parameters
    ----------
    symbol : Symbol
        The symbol configuration of computation network.

    ctx : Context or list of Context, optional
        The device context of training and prediction.
        To use multi GPU training, pass in a list of gpu contexts.

    num_epoch : int, optional
        Training parameter, number of training epochs(epochs).

    optimizer : str or Optimizer, optional
        Training parameter, name or optimizer object for training.

    initializier : initializer function, optional
        Training parameter, the initialization scheme used.

    numpy_batch_size : int, optional
        The batch size of training data.
        Only needed when input array is numpy.

    arg_params : dict of str to NDArray, optional
        Model parameter, dict of name to NDArray of net's weights.

    aux_params : dict of str to NDArray, optional
        Model parameter, dict of name to NDArray of net's auxiliary states.

    allow_extra_params : boolean, optional
        Whether allow extra parameters that are not needed by symbol
        to be passed by aux_params and arg_params.
        If this is True, no error will be thrown when aux_params and arg_params
        contain extra parameters than needed.

    begin_epoch : int,optional
        The begining training epoch.

    **kwargs : dict
        The additional keyword arguments passed to optimizer.
    """
    def __init__(self, symbol, ctx=None,
                 num_epoch=None, optimizer='sgd',
                 initializer=Uniform(0.01),
                 numpy_batch_size=128,
                 arg_params=None, aux_params=None,
                 allow_extra_params=False,
                 begin_epoch=0,
                 **kwargs):
        # check if symbol contain duplicated names.
        _check_arguments(symbol)
        # rematch parameters to delete useless ones
        if allow_extra_params:
            if arg_params:
                arg_names = set(symbol.list_arguments())
                arg_params = {k : v for k, v in arg_params.items()
                              if k in arg_names}
            if aux_params:
                aux_names = set(symbol.list_auxiliary_states())
                aux_params = {k : v for k, v in aux_params.items()
                              if k in aux_names}
        # basic configuration
        self.symbol = symbol
        if ctx is None:
            ctx = [cpu()]
        elif isinstance(ctx, Context):
            ctx = [ctx]
        self.ctx = ctx
        # training parameters
        self.num_epoch = num_epoch
        self.kwargs = kwargs.copy()
        self.optimizer = optimizer
        self.initializer = initializer
        self.numpy_batch_size = numpy_batch_size
        # model parameters
        self.arg_params = arg_params
        self.aux_params = aux_params
        # internal helper state
        self._pred_exec = None
        self._pred_exec_input = None
        self.begin_epoch = begin_epoch

    @staticmethod
    def _is_data_arg(name):
        """Check if name is a data argument."""
        return name.endswith('data') or name.endswith('label')

    def _init_params(self, input_shape):
        """Use initializer to initialize the parameters."""
        arg_shapes, _, aux_shapes = self.symbol.infer_shape(data=input_shape)
        if self.arg_params is None:
            arg_names = self.symbol.list_arguments()
            self.arg_params = {k : nd.zeros(s) for k, s in list(zip(arg_names, arg_shapes))
                               if not self._is_data_arg(k)}
        if self.aux_params is None:
            aux_names = self.symbol.list_auxiliary_states()
            self.aux_params = {k : nd.zeros(s) for k, s in list(zip(aux_names, aux_shapes))}
        for k, v in self.arg_params.items():
            self.initializer(k, v)
        for k, v in self.aux_params.items():
            self.initializer(k, v)

    def __getstate__(self):
        this = self.__dict__.copy()
        this['_pred_exec'] = None
        return this

    def __setstate__(self, state):
        self.__dict__.update(state)

    def _init_predictor(self, input_shape):
        """Initialize the predictor module for running prediction."""
        if self._pred_exec is not None:
            return
        # for now only use the first device
        pred_exec = self.symbol.simple_bind(
            self.ctx[0], grad_req='null', data=input_shape)

        for name, value in list(zip(self.symbol.list_arguments(), pred_exec.arg_arrays)):
            if not self._is_data_arg(name):
                if not name in self.arg_params:
                    raise ValueError("%s not exist in arg_params" % name)
                self.arg_params[name].copyto(value)
        for name, value in list(zip(self.symbol.list_auxiliary_states(), pred_exec.aux_arrays)):
            assert name in self.aux_params
            self.aux_params[name].copyto(value)
        data_index, _ = _check_arguments(self.symbol)
        self._pred_exec_input = pred_exec.arg_arrays[data_index]
        self._pred_exec = pred_exec

    def _init_iter(self, X, y, is_train):
        """Initialize the iterator given input."""
        if isinstance(X, (np.ndarray, nd.NDArray)):
            if y is None:
                if is_train:
                    raise ValueError('y must be specified when X is numpy.ndarray')
                else:
                    y = np.zeros(X.shape[0])
            if not isinstance(y, (np.ndarray, nd.NDArray)):
                raise TypeError('y must be ndarray when X is numpy.ndarray')
            if X.shape[0] != y.shape[0]:
                raise ValueError("The numbers of data points and labels not equal")
            if y.ndim == 2 and y.shape[1] == 1:
                y = y.flatten()
            if y.ndim != 1:
                raise ValueError("Label must be 1D or 2D (with 2nd dimension being 1)")
            return io.NDArrayIter(X, y, self.numpy_batch_size, shuffle=is_train)
        if not isinstance(X, io.DataIter):
            raise TypeError('X must be DataIter, NDArray or numpy.ndarray')
        return X

    def _init_eval_iter(self, eval_data):
        """Initialize the iterator given eval_data."""
        if eval_data is None:
            return eval_data
        if isinstance(eval_data, (tuple, list)) and len(eval_data) == 2:
            if eval_data[0] is not None:
                if eval_data[1] is None and isinstance(eval_data[0], io.DataIter):
                    return eval_data[0]
                input_data = (np.array(eval_data[0]) if isinstance(eval_data[0], list)
                              else eval_data[0])
                input_label = (np.array(eval_data[1]) if isinstance(eval_data[1], list)
                               else eval_data[1])
                return self._init_iter(input_data, input_label, is_train=True)
            else:
                raise ValueError("Eval data is NONE")
        if not isinstance(eval_data, io.DataIter):
            raise TypeError('Eval data must be DataIter, or ' \
                            'NDArray/numpy.ndarray/list pair (i.e. tuple/list of length 2)')
        return eval_data

    def predict(self, X):
        """Run the prediction, always only use one device.

        Parameters
        ----------
        X : mxnet.DataIter or numpy.ndarray

        Returns
        -------
        y : numpy.ndarray
            The predicted value of the output.
        """
        X = self._init_iter(X, None, is_train=False)
        X.reset()
        data, _ = X.next()
        self._init_predictor(data.shape)

        outputs = []
        X.reset()
        for data, _ in X:
            data.copyto(self._pred_exec_input)
            self._pred_exec.forward(is_train=False)
            out_batch = self._pred_exec.outputs[0].asnumpy()
            padded = X.getpad()
            real_size = out_batch.shape[0] - padded
            out_batch = out_batch[0:real_size, :]
            outputs.append(out_batch)
        return np.concatenate(outputs)

    def fit(self, X, y=None, eval_data=None, eval_metric='acc',
            epoch_end_callback=None, batch_end_callback=None,
            kvstore='local', logger=None):
        """Fit the model.

        Parameters
        ----------
        X : DataIter, or numpy.ndarray/NDArray
            Training data.

        y : numpy.ndarray/NDArray, optional
            Training set label.
            If X is numpy.ndarray/NDArray, y is required to be set.
            While y can be 1D or 2D (with 2nd dimension as 1), its 1st dimension must be
                the same as X, i.e. the number of data points and labels should be equal.

        eval_data : DataIter or numpy.ndarray/list/NDArray pair
            If eval_data is numpy.ndarray/list/NDArray pair,
                it should be (valid_data, valid_label).

        eval_metric : metric.EvalMetric or str or callable
            The evaluation metric, name of evaluation metric.
            Or a customize evaluation function that returns the statistics
            based on minibatch.

        epoch_end_callback : callable(epoch, symbol, arg_params, aux_states)
            A callback that is invoked at end of each epoch.
            This can be used to checkpoint model each epoch.

        batch_end_callback: callable(epoch)
            A callback that is invoked at end of each batch
            For print purpose

        kvstore: KVStore or str, optional
           The KVStore or a string kvstore type:
           'local' : multi-devices on a single machine, will automatically
               choose one from 'local_update_cpu', 'local_allreduce_cpu', and
              'local_allreduce_device'
           'dist_sync' : multi-machines with BSP
           'dist_async' : multi-machines with partical asynchronous

           In default uses 'local', often no need to change for single machiine.

        logger : logging logger, optional
            When not specified, default logger will be used.

        """
        X = self._init_iter(X, y, is_train=True)
        eval_data = self._init_eval_iter(eval_data)
        # Simply ignore the first example to get input_shape
        # in first training epoch.
        if not X.iter_next():
            X.reset()
            assert X.iter_next()
        input_shape = X.getdata().shape
        if self.arg_params is None:
            self._init_params(input_shape)

        # setup metric
        if not isinstance(eval_metric, metric.EvalMetric):
            eval_metric = metric.create(eval_metric)

        # create kvstore
        (kvstore, update_on_kvstore) = _create_kvstore(
            kvstore, len(self.ctx), self.arg_params)

        # init optmizer
        if isinstance(self.optimizer, str):
            batch_size = input_shape[0]
            if kvstore and kvstore.type == 'dist_sync':
                batch_size *= kvstore.num_workers

        optimizer = opt.create(self.optimizer,
                               rescale_grad=(1.0/batch_size),
                               **(self.kwargs))
        # do training
        _train_multi_device(self.symbol, self.ctx, input_shape,
                            self.arg_params, self.aux_params,
                            begin_epoch=self.begin_epoch, end_epoch=self.num_epoch,
                            optimizer=optimizer,
                            train_data=X, eval_data=eval_data,
                            eval_metric=eval_metric,
                            epoch_end_callback=epoch_end_callback,
                            batch_end_callback=batch_end_callback,
                            kvstore=kvstore, update_on_kvstore=update_on_kvstore,
                            logger=logger)

    def save(self, prefix, epoch=None):
        """Checkpoint the model checkpoint into file.

        You can also use pickle to do the job if you only work on python.
        The advantage of load/save is the file is language agnostic.
        This means the file saved using save can be loaded by other language binding of mxnet.
        You also get the benefit being able to directly load/save from cloud storage(S3, HDFS)

        Parameters
        ----------
        prefix : str
            Prefix of model name.

        See Also
        --------
        Symbol.load : the method to load the model back.

        Notes
        -----
        - ``prefix-symbol.json`` will be saved for symbol.
        - ``prefix-epoch.params`` will be saved for parameters.
        """
        if epoch is None:
            epoch = self.num_epoch
        assert epoch is not None
        save_checkpoint(prefix, epoch, self.symbol, self.arg_params, self.aux_params)

    @staticmethod
    def load(prefix, epoch, ctx=None, **kwargs):
        """Load model checkpoint from file.

        Parameters
        ----------
        prefix : str
            Prefix of model name.

        epoch : int
            epoch number of model we would like to load.

        ctx : Context or list of Context, optional
            The device context of training and prediction.
        kwargs : dict
            other parameters for model, including num_epoch, optimizer and numpy_batch_size

        Returns
        -------
        model : FeedForward
            The loaded model that can be used for prediction.

        Notes
        -----
        - ``prefix-symbol.json`` will be saved for symbol.
        - ``prefix-epoch.params`` will be saved for parameters.
        """
        symbol, arg_params, aux_params = load_checkpoint(prefix, epoch)
        return FeedForward(symbol, ctx=ctx,
                           arg_params=arg_params, aux_params=aux_params,
                           begin_epoch=epoch,
                           **kwargs)

    @staticmethod
    def create(symbol, X, y=None, ctx=None,
               num_epoch=None, optimizer='sgd', initializer=Uniform(0.01),
               eval_data=None, eval_metric='acc', epoch_end_callback=None,
               kvstore='local', logger=None, **kwargs):
        """Functional style to create a model.

        This function will be more consistent with functional
        languages such as R, where mutation is not allowed.

        Parameters
        ----------
        symbol : Symbol
            The symbol configuration of computation network.

        X : DataIter
            Training data

        y : numpy.ndarray, optional
            If X is numpy.ndarray y is required to set

        ctx : Context or list of Context, optional
            The device context of training and prediction.
            To use multi GPU training, pass in a list of gpu contexts.

        num_epoch : int, optional
            Training parameter, number of training epochs(epochs).

        optimizer : str or Optimizer, optional
            Training parameter, name or optimizer object for training.

        initializier : initializer function, optional
            Training parameter, the initialization scheme used.

        eval_data : DataIter or numpy.ndarray pair
            If eval_set is numpy.ndarray pair, it should be (valid_data, valid_label)

        eval_metric : metric.EvalMetric or str or callable
            The evaluation metric, name of evaluation metric.
            Or a customize evaluation function that returns the statistics
            based on minibatch.

        epoch_end_callback : callable(epoch, symbol, arg_params, aux_states)
            A callback that is invoked at end of each epoch.
            This can be used to checkpoint model each epoch.

        kvstore: KVStore or str, optional
           The KVStore or a string kvstore type:
           'local' : multi-devices on a single machine, will automatically
               choose one from 'local_update_cpu', 'local_allreduce_cpu', and
              'local_allreduce_device'
           'dist_sync' : multi-machines with BSP
           'dist_async' : multi-machines with partical asynchronous

           In default uses 'local', often no need to change for single machiine.
        """
        model = FeedForward(symbol, ctx=ctx, num_epoch=num_epoch,
                            optimizer=optimizer, initializer=initializer, **kwargs)
        model.fit(X, y, eval_data=eval_data, eval_metric=eval_metric,
                  epoch_end_callback=epoch_end_callback,
                  kvstore=kvstore,
                  logger=logger)
        return model
