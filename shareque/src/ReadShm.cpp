#include <iostream>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <sys/shm.h>
#include "ShareQueue.h"

using namespace std;

const char *pShmPath = "/tmp/Shm";
const char *pSemPath = "/tmp/SemMtx";

int main()
{
  CShareQueue shq(pShmPath, pSemPath, 0, 128);
  if (shq.Initialize() < 0)
  {
    ERROR_LOG("Shm Initialize Error!");
    return 0;
  }
  int count = 0;
  int ret = 0;
  while (1)
  {
    DEBUG_LOG("count:" << count);
    count++;
    shq.ShmPrint();
    if (!shq.Empty())
    {
      int element;
      if ((ret = shq.PopOut(&element)) < 0)
      {
        ERROR_LOG("ret = " << ret);
        DEBUG_LOG(std::string(shq.ShmErrorMsgBuff));
        break;
      }
      DEBUG_LOG("element:" << element);
      if (element == -3)
      {
        break;
      }
    }
    else
    {
      // sleep(1);
    }
    sleep(1);
  }
  return 0;
}
