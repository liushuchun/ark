# for dynamic linking with MKL
#CFLAGS=-I${MKL_INC} -L${MKL_LIB} -Wl,--start-group -lmkl_intel_lp64 -lmkl_intel_thread -lmkl_core -Wl,--end-group -openmp -lpthread 

#MKL_LIB=/opt/intel/mkl/lib/intel64
MKL_LIB=/opt/intel/compilers_and_libraries/linux/mkl/lib/intel64
COMPILER_LIB=/opt/intel/compilers_and_libraries/linux/lib/intel64
# for static linking
#mkl=-O3 -Wl,--start-group $(MKL_LIB)/libmkl_intel_lp64.a $(MKL_LIB)/libmkl_intel_thread.a $(MKL_LIB)/libmkl_core.a -Wl,--end-group -I$(MKL_INC) -openmp -lpthread 
mkl=-std=c99 -DMKL_DIRECT_CALL -Wl,--start-group ${MKL_LIB}/libmkl_intel_lp64.a ${COMPILER_LIB}/libiomp5.a ${MKL_LIB}/libmkl_core.a ${MKL_LIB}/libmkl_intel_thread.a --end-group -lpthread
mklnew=-std=c99 -DMKL_DIRECT_CALL -Wl,--start-group /root/lnx/mkl/lib/intel64/libmkl_intel_lp64.a /root/lnx/compiler/lib/intel64/libiomp5.a /root/lnx/mkl/lib/intel64/libmkl_core.a /root/lnx/mkl/lib/intel64/libmkl_intel_thread.a --end-group -lpthread

all: sgemm-static sgemm-new
#xtest : sgemm.c
#	icc $< -o xtest ${mkl} 

#export KMP_AFFINITY=granularity=fine,compact,1,0
#icc -o sgemm sgemm.c -mkl -std=c99
#gcc -o sgemm-blas-hsw sgemm-blas.c -I ~/OpenBLAS-sandybridge ~/OpenBLAS-sandybridge/libopenblas.a -lpthread -L ~/OpenBLAS-sandybridge
#gcc -o sgemm-openblas sgemm-blas.c -I /root/openblas/OpenBLAS /root/openblas/OpenBLAS/libopenblas.a -lpthread -L /root/openblas/OpenBLAS
#
#static
sgemm-static: sgemm-static.c
	icc $< -o sgemm-static -O3 ${mkl}

sgemm-new: sgemm-static.c
	icc $< -o sgemm-new -O3 ${mklnew}
#icc -o sgemm-static -g -O3 sgemm-static.c -std=c99 -Wl,--start-group /opt/intel/mkl/lib/intel64/libmkl_intel_lp64.a /opt/intel/lib/intel64/libiomp5.a /opt/intel/mkl/lib/intel64/libmkl_core.a /opt/intel/mkl/lib/intel64/libmkl_intel_thread.a --end-group -lpthread

clean:
	rm -f sgemm-static sgemm-new


