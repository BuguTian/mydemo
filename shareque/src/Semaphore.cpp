#include "Semaphore.h"

CSemaphore::CSemaphore(const char *pSemPath)
{
  if (Init(pSemPath) < 0)
  {
    DEBUG_LOG("CSemaphore initilize error!");
  }
}
CSemaphore::~CSemaphore()
{
  if (semctl(iSemId, 0, IPC_RMID) == -1)
  {
    DEBUG_LOG("Semctl Err");
  }
}
int CSemaphore::Init(const char *pSemPath)
{
  if ((iSemKey = ftok(pSemPath, 1)) == -1)
  {
    DEBUG_LOG("Sem Ftok Err");
    return -1;
  }
  if ((iSemId = semget(iSemKey, 1, IPC_CREAT | 0666)) == -1)
  {
    DEBUG_LOG("Semget Err");
    return -1;
  }

  UnLock();
  return 0;
}
void CSemaphore::P()
{
  sSem.sem_op = -1;
  // UNDO防止进程意外退出造成其他信号量永久阻塞
  sSem.sem_flg = SEM_UNDO;
  semop(iSemId, &sSem, 1);
}
void CSemaphore::V()
{
  sSem.sem_op = 1;
  sSem.sem_flg = SEM_UNDO;
  semop(iSemId, &sSem, 1);
}
void CSemaphore::Lock()
{
  sSem.sem_num = 0;
  sSem.sem_op = -1;
  // UNDO防止进程意外退出造成其他信号量永久阻塞
  sSem.sem_flg = SEM_UNDO;
  semop(iSemId, &sSem, 1);
}
void CSemaphore::UnLock()
{
  sSem.sem_num = 0;
  sSem.sem_op = 1;
  sSem.sem_flg = SEM_UNDO;
  semop(iSemId, &sSem, 1);
}
int CSemaphore::TryLock()
{
  return sSem.sem_num;
}
