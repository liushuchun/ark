FROM nvidia/cuda:8.0-cudnn5-devel
MAINTAINER liushuchun <liusc@gmail.com>

RUN sed -i s/archive.ubuntu.com/mirrors.163.com/g /etc/apt/sources.list
RUN sed -i s/security.ubuntu.com/mirrors.163.com/g /etc/apt/sources.list
# 这两个 NVIDIA source list 更新存在问
RUN rm /etc/apt/sources.list.d/cuda.list /etc/apt/sources.list.d/nvidia-ml.list


RUN apt-get update && apt-get install -y \
    build-essential git libatlas-base-dev  \
    libcurl4-openssl-dev libgtest-dev cmake wget unzip net-tools  python-dev python3-dev liblapacke-dev libopenblas-dev vim liblapacke-dev checkinstall  openssh-server ssh  ca-certificates   lrzsz curl  unzip  cmake \
    python-dev python-pip python-tk libopenblas-dev \
    python-setuptools


RUN cd /usr/src/gtest && cmake CMakeLists.txt && make && cp *.a /usr/lib

RUN apt-get update && apt-get install -y python-dev python3-dev



#python package
RUN pip install -U pip setuptools && pip install nose pylint numpy nose-timer requests


#run ssh-server
RUN service ssh start



# Install git, wget and other dependencies
RUN apt-get update && apt-get install -y \
  git \
  python-numpy  pkg-config \
  wget

# Install Opencv into the workspace
RUN mkdir /workspace && mkdir /workspace/opencv && mkdir /workspace/opencv-contrib

RUN export OPENCV_CONTRIB_ROOT=/workspace/opencv-contrib OPENCV_ROOT=/workspace/opencv OPENCV_VER=3.2.0 && \
    git clone -b ${OPENCV_VER} --depth 1 https://github.com/opencv/opencv.git ${OPENCV_ROOT} && \
    git clone -b ${OPENCV_VER} --depth 1 https://github.com/opencv/opencv_contrib.git ${OPENCV_CONTRIB_ROOT} && \
    mkdir -p ${OPENCV_ROOT}/build && cd ${OPENCV_ROOT}/build && \
    cmake -D CMAKE_BUILD_TYPE=RELEASE -D CMAKE_INSTALL_PREFIX=/usr/local \
    -D INSTALL_C_EXAMPLES=OFF -D INSTALL_PYTHON_EXAMPLES=OFF \
    -D OPENCV_EXTRA_MODULES_PATH=${OPENCV_CONTRIB_ROOT}/modules \
    -D WITH_CUDA=ON -D BUILD_opencv_python2=ON -D BUILD_EXAMPLES=OFF .. && \
    make -j16 && make install && ldconfig

#environment params
ENV LD_LIBRARY_PATH=/usr/local/cuda-8.0/lib64/
ENV PYTHONPATH /workspace/mxnet/python:$PYFCIS_ROOT:$PYTHONPATH
ENV MXNET_ROOT=/workspace/mxnet
ENV PATH /usr/local/cuda/bin:$PATH
ENV MXNET_CUDNN_AUTOTUNE_DEFAULT=1

RUN cd /workspace && git clone --recursive https://github.com/apache/incubator-mxnet.git mxnet &&   cd /workspace/mxnet && make -j16 USE_CUDA=1 USE_CUDA_PATH=/usr/local/cuda-8.0 USE_CUDNN=1 USE_OPENCV=1 USE_DIST_KVSTORE=1 USE_BLAS=openblas


ENV PYTHONPATH ${MXNET_ROOT}/python:$PYFCIS_ROOT:$PYTHONPATH

# Install Python package
RUN cd /workspace/mxnet/python && python setup.py install

WORKDIR /workspace/
