#bash
sudo apt-get update &&
sudo apt-get -y install build-essential git cmake &&
sudo apt-get -y install libprotobuf-dev libleveldb-dev libsnappy-dev &&
sudo apt-get -y install libopencv-dev libhdf5-serial-dev protobuf-compiler &&
sudo apt-get -y install --no-install-recommends libboost-all-dev &&
sudo apt-get -y install libgflags-dev libgoogle-glog-dev liblmdb-dev &&
sudo apt-get -y install libatlas-base-dev

current="$pwd"



find .-type f -exec sed -i -e 's^"hdf5.h"^"hdf5/serial/hdf5.h"^g' -e 's^"hdf5_hl.h"^"hdf5/serial/hdf5_hl.h"^g' '{}' ;
cd /usr/lib/x86_64-linux-gnu
sudo ln -s libhdf5_serial.so.10.1.0 libhdf5.so
sudo ln -s libhdf5_serial_hl.so.10.0.2 libhdf5_hl.so


echo 'source /opt/intel/bin/compilervars.sh intel64' >> ~/.bashrc

cd ~
git clone https://github.com/intel/caffe.git


cd caffe

cp $current/Makefile.config ./

echo "export CAFFE_ROOT=`pwd`" >> ~/.bashrc
source ~/.bashrc

cd /usr/lib/x86_64-linux-gnu
sudo ln -s libhdf5_serial.so.10.1.0 libhdf5.so
sudo ln -s libhdf5_serial_hl.so.10.0.2 libhdf5_hl.so


cd ~/caffe

sudo make -j50

#install python relative
sudo apt-get -y install gfortran python-dev python-pip

cd ~/caffe/python
for req in $(cat requirements.txt); do sudo pip install $req; done

sudo pip install scikit-image #depends on other packages

sudo ln -s /usr/include/python2.7/ /usr/local/include/python2.7

sudo ln -s /usr/local/lib/python2.7/dist-packages/numpy/core/include/numpy/ \
  /usr/local/include/python2.7/numpy
cd ~/caffe

make pycaffe -j NUM_THREADS

echo "export PYTHONPATH=$CAFFE_ROOT/python" >> ~/.bashrc

source ~/.bashrc








