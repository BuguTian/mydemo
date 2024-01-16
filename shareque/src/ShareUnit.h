#ifndef __SHARE_UNIT_H__
#define __SHARE_UNIT_H__

#include <iostream>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/shm.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/sem.h>
#include <typeinfo>

#include "Log.h"
#include "Semaphore.h"

using namespace std;

// 队中每个元素前面插入起始标志2；
// 队中每个元素后面插入结束标志3；
// 队列本身并不用适配类型，完全用二进制的方式传输
typedef struct ShareUnit
{
  int crcCode;
  int len;
  char content[0];
};
#endif // __SHARE_UNIT_H__
