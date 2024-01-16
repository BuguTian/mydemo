#ifndef __SEMAPHORE_H__
#define __SEMAPHORE_H__

#include <iostream>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/shm.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/sem.h>

#include "Log.h"

using namespace std;

class CSemaphore
{
public:
    CSemaphore(const char *pSemPath);
    ~CSemaphore();
    int Init(const char *pSemPath);
    void P();
    void V();
    void Lock();
    void UnLock();
    int TryLock();
private:
    key_t iSemKey;
    int iSemId;
    struct sembuf sSem;
};
#endif // __SEMAPHORE_H__
