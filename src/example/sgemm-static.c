#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include <mkl.h>
#include <omp.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/shm.h>
#include <ctype.h>
#include <fcntl.h>           /* For O_* constants */
#include <sys/stat.h>        /* For mode constants */
#include <semaphore.h>
#include <unistd.h>


#define FLOAT_TYPE float
#define DATA_LEN 524288

int main(int argc, char* argv[])
{
    FLOAT_TYPE* A;
	FLOAT_TYPE* B;
	FLOAT_TYPE* C;
    FLOAT_TYPE  alpha = 1.0, beta = 1.0;

	int m = 1024;
	int n = 1024;
	int k = 1024;
    int LOOP_COUNT = 10;
    int i, cores = 1, target = 1, start = 1;
    int fpc = 32;  // default hsw 2.3GHz, FMA enabled. fpc: float operating per cycle
    float hz = 2.3; // default hsw 2.3GHz, FMA enabled. fpc: float operating per cycle
   
    if (argc == 1) { 
        printf("Usage: %s m n k loop cores hz fpc\n", argv[0]);
        printf("\t hz: CPU frequency in GHz, default=2.3\n");
        printf("\t fpc: float ops per cycle, hsw=32, snb/ivb=16, default=32\n");
    }

	if (argc >= 2) m = atoi(argv[1]);
	if (argc >= 3) n = atoi(argv[2]);
	if (argc >= 4) k = atoi(argv[3]);
	if (argc >= 5) LOOP_COUNT = atoi(argv[4]);
    if (argc >= 6) cores = atoi(argv[5]); 
    if (argc >= 7) hz = (float)atof(argv[6]); 
    if (argc >= 8) fpc = atoi(argv[7]); 

	char transa='N';
	char transb='N';

	A = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE)*m*k, 32);
	B = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE)*k*n, 32);
	C = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE)*m*n, 32);
	float * xdata = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE)*256, 32);
	float * ydata = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE)*2048, 32);
	float * mdata = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE)*256*2048, 32);
        float * data = (FLOAT_TYPE*) mkl_malloc(sizeof(FLOAT_TYPE) * DATA_LEN, 32);

	for (i=0; i<m*k ; ++i)
	{
		A[i] = (FLOAT_TYPE)rand() / RAND_MAX;
	}

	for (i=0; i<k*n ; ++i)
	{
		B[i] = (FLOAT_TYPE)rand() / RAND_MAX;
	}
    
	double gflop = (2.0*m*n*k)*1E-9;


   {   
    /* first call for thread/buffer initialization */
        omp_set_num_threads(cores); // set thread numbers
        sgemm(&transa, &transb, &m, &n, &k, &alpha, A, &m, B, &k, &beta, C, &m);

	    double time_st  = dsecnd();
        for (i=0; i<LOOP_COUNT; ++i)
        {
            sgemm(&transa, &transb, &m, &n, &k, &alpha, A, &m, B, &k, &beta, C, &m);
        }
	    double time_end = dsecnd() ;
        double time_avg = (time_end - time_st)/LOOP_COUNT;
        printf ("m=%d,n=%d,k=%d cores=%d gflop=%.5f peak=%.5f efficiency=%.5f\n", m, n, k, cores, gflop / time_avg, (hz * fpc * cores), (gflop / time_avg) / (hz * fpc * cores));
    }

	mkl_free(A);
	mkl_free(B);
	mkl_free(C);

	return 0;
 }
