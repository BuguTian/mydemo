#ifndef __SHARE_QUEUE_H__
#define __SHARE_QUEUE_H__

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
#include "ErrorCode.h"
#include "Semaphore.h"

using namespace std;

// 虚拟的队列结构，存储队列信息
typedef struct VtlQueue
{
  int rear;              // 队尾索引
  int front;             // 队头索引
  int msgNum;            // 队列中现存多少个消息
  int flag;              // front < rear ? 0 : 1,
  int buffSize;          // 通常是m_iMaxShmSize - 40
  long long lastPopTime; // 上一次出队时间
  char dataBuff[0];      // 可变长数组,存队列内容
} VtlQueue;

// 每一帧消息的描述结构
typedef struct MsgFrame
{
  int crcCode;         // Crc校验码
  int len;             // 消息体的长度
  long long timeStamp; // 插入时的时间戳，用于到时候用超时删除任务
  char content[0];     // 预留，暂时不做处理
} MsgFrame;

class CShareQueue
{
public:
  /*初始化的时候需要指定
   *1.共享内存的路径
   *2.用于对共享内存同步的信号量的路径
   *3.标识此对象是共享内存的读端还是写端
   */
  CShareQueue(const char *pShmPath, const char *pMtxPath, int flag, int iMaxShmSize = 4096);
  ~CShareQueue();
  int Initialize();

  int GetFreeSize();                                      // 获取当前队列的空闲大小
  int FindMsgFrame(int &idxbegin, int &idxend, int &len); // 找到head+msgFrame+data+tail的起止位置和长度
  int CheckMsgFrame(const int &len);                      // 校验数据格式是否正确

  int EnQue(const void *src, int len);                  // 入队操作
  int DeQue(void *dst, int len);                        // 出队操作
  int FtQue(void *dst, int len);                        // 取队头元素
  int DlQue(int len);                                   // 删除队头元素
  int UndoQue(int len);                                 // 删除刚入队的元素
  int GetDataWithIdx(void *dst, int idxbegin, int len); // 取某段(idxbegin)索引之后的len字节的数据

  // 下面是对上面函数的封装，增加了同步操作,上面的函数都没有做同步操作
  int Push(const void *src, int len); // force：是否强制入队，0：不强制，如果队满就返回-1,1：强制入队，覆盖队头元素,默认不强制
  int Front(void *dst);
  int Pop();                           // 不建议先Front后Pop的做法，因为涉及到异步操作，有可能你出队的元素不是你已经读取的元素
  int PopOut(void *dst, int flag = 2); // flag：0表示只复制出数据并不出队，1表示只出队，2表示出队同时拷贝出数据
  int Size();
  bool Empty();
  bool Full();
  int GetMsgNum();
  long long GetLastPopTime();
  void DeleteOuttimeMsg(long long referTime);
  virtual void ShmPrint(); // 遍历打印当前共享内存状态，需要根据类型自己实现,默认用16进制输出

public:
  char ShmErrorMsgBuff[2048];

protected:
  VtlQueue *m_vtlQueue;
  int m_iShmKey;
  int m_iShmId;
  int m_iFlag;        // 0:表示读，1:表示写
  int m_iMaxShmSize;  // 共享内存的总大小
  int m_iMaxBuffSize; // 共享结构体中buffer的大小

  CSemaphore m_rwMtx;

  // int m_ifBlock;     //条件变量与读写同步

  std::string m_strShmPath;
  std::string m_strRWSemPath;
  std::string m_strCondPath;
};

#endif // __SHARE_QUEUE_H__
