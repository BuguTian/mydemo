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
  CShareQueue shq(pShmPath, pSemPath, 1, 128);
  if (shq.Initialize() < 0)
  {
    DEBUG_LOG("Shm Initialize Error!");
    return 0;
  }
  int i = 0;
  int ret = 0;
  long long start = Milliseconds();
  for (i = 0; i < 100;)
  {
    DEBUG_LOG("i:" << i);
    if ((ret = shq.Push(&i, sizeof(int))) >= 0)
    {
      i++;
    }
    else
    {
      DEBUG_LOG(std::string(shq.ShmErrorMsgBuff));
    }
    DEBUG_LOG("ret=" << ret);
    shq.ShmPrint();
    sleep(1);
  }
  // Push -3 是为了通知对方退出,此值由双方约定
  int endflag = -3;
  shq.Push(&endflag, sizeof(int));
  shq.ShmPrint();
  long long end = Milliseconds();
  // 耗时约63417
  DEBUG_LOG("It takes " << end - start << " milliseconds.");
  return 0;
}
