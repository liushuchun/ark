apt-get update && apt-get install -y \
    build-essential git libatlas-base-dev  \
    libcurl4-openssl-dev libgtest-dev cmake wget unzip net-tools  python-dev python3-dev liblapacke-dev libopenblas-dev vim liblapacke-dev checkinstall graphviz openssh-server ssh

cd /usr/src/gtest && cmake CMakeLists.txt && make && cp *.a /usr/lib



cd /tmp && wget https://bootstrap.pypa.io/get-pip.py && python3 get-pip.py && python2 get-pip.py

pip2 install nose pylint numpy nose-timer requests



echo "LD_LIBRARY_PATH=/usr/local/cuda-8.0/lib64/:$LD_LIBRARY_PA" >> ~/.bashrc
source ~/.bashrc

apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates wget vim lrzsz curl git unzip build-essential cmake \
    python-dev python-pip python-tk libopenblas-dev \
    libatlas-base-dev libcurl4-openssl-dev \
    libgtest-dev python-setuptools python-numpy \
    openssh-server rsync && \
    cd /usr/src/gtest && cmake CMakeLists.txt && make && cp *.a /usr/lib && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install git, wget and other dependencies
service ssh start
mkdir /workspace/opencv
mkdir /workspace/opencv-contrib

export OPENCV_CONTRIB_ROOT=/workspace/opencv-contrib OPENCV_ROOT=/workspace/opencv OPENCV_VER=3.2.0 && \
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
    -D WITH_CUDA=ON -D BUILD_opencv_python2=ON -D BUILD_EXAMPLES=OFF .. && \
    make -j16 && make install && ldconfig


export PYTHONPATH=/workspace/mxnet/python

cd /workspace && git clone --recursive https://github.com/dmlc/mxnet && cd mxnet && \
    make -j8 USE_CUDA=1 USE_CUDA_PATH=/usr/local/cuda-8.0 USE_CUDNN=1 USE_DIST_KVSTORE=1 USE_BLAS=openblas

# Install Python package
cd /workspace/mxnet/python && python setup.py install

