#include "ShareQueue.h"

CShareQueue::CShareQueue(const char *pShmPath, const char *pMtxPath,
                         int flag, int iMaxShmSize) : m_rwMtx(pMtxPath)
{
  m_strShmPath = std::string(pShmPath);
  m_strRWSemPath = std::string(pMtxPath);
  m_iFlag = flag;
  m_iMaxShmSize = iMaxShmSize;
  m_iMaxBuffSize = iMaxShmSize - sizeof(VtlQueue) - 8;
  /*if (Initialize() < 0)
  {
      ERROR_LOG("Initialize Err!");
  }*/
}

CShareQueue::~CShareQueue()
{
  if (shmdt((void *)m_vtlQueue) == -1)
  {
    ERROR_LOG("shmdt failed!");
    sprintf(ShmErrorMsgBuff, "shmdt failed!\n");
  }
  // 最后的共享内存由读端删除
  // 什么时候删除，要看读写双方约定的特殊element
  if ((!m_iFlag) && shmctl(m_iShmId, IPC_RMID, NULL) == -1)
  {
    ERROR_LOG("Shmctl Error!");
    sprintf(ShmErrorMsgBuff, "shmdt failed!");
  }
}

int CShareQueue::Initialize()
{
  int ret = 0;
  memset(ShmErrorMsgBuff, 0, sizeof(ShmErrorMsgBuff));
  // 通过双方都约定的路径，获取shm key
  if ((m_iShmKey = ftok(m_strShmPath.c_str(), 1)) == -1)
  {
    ERROR_LOG("Shm Ftok Error!");
    sprintf(ShmErrorMsgBuff, "Shm Ftok Error! ErrorCode=%d\n", ERRCODE_QUE_INIT_FAILED);
    return -1;
  }
  // shmflg：0为获取，IPC_CREAT为不存在则创建，IPC_CREAT|IPC_EXCL为必须不存在，然后由它来创建，否则返回错误-1
  if ((m_iShmId = shmget(m_iShmKey, m_iMaxShmSize, IPC_CREAT | 0666)) == -1)
  {
    ERROR_LOG("Shmget Error!");
    sprintf(ShmErrorMsgBuff, "Shmget Error! ErrorCode=%d\n", ERRCODE_QUE_INIT_FAILED);
    return -1;
  }
  char *tmpPtr;
  if ((tmpPtr = (char *)shmat(m_iShmId, NULL, 0)) == (char *)-1)
  {
    ERROR_LOG("Shmat Error!");
    sprintf(ShmErrorMsgBuff, "Shmat Error! ErrorCode=%d\n", ERRCODE_QUE_INIT_FAILED);
    return -1;
  }
  m_vtlQueue = (VtlQueue *)tmpPtr;
  m_rwMtx.Lock();
  m_vtlQueue->buffSize = m_iMaxBuffSize;
  if (m_vtlQueue->msgNum < 0 || m_vtlQueue->msgNum > m_iMaxBuffSize)
  {
    m_vtlQueue->msgNum = 0;
  }
  if (!(m_vtlQueue->front > 0 && m_vtlQueue->front < m_vtlQueue->buffSize && m_vtlQueue->rear > 0 && m_vtlQueue->rear < m_vtlQueue->buffSize))
  {
    memset(m_vtlQueue->dataBuff, 0, m_vtlQueue->buffSize);
    m_vtlQueue->front = m_vtlQueue->rear = 0;
    m_vtlQueue->flag = 1;
    m_vtlQueue->msgNum = 0;
  }
  m_rwMtx.UnLock();

  return ret;
}

int CShareQueue::EnQue(const void *src, int len)
{
  if (m_vtlQueue->flag == 1)
  {
    int need = m_vtlQueue->front + len - m_vtlQueue->buffSize;
    if (need > 0)
    {
      if (need > m_vtlQueue->rear)
      {
        return -1;
      }
      int sgpt = len - need;
      memcpy((m_vtlQueue->dataBuff + m_vtlQueue->front), src, sgpt);
      memcpy(m_vtlQueue->dataBuff, ((char *)src + sgpt), need);
      m_vtlQueue->front = need;
      m_vtlQueue->flag = 0;
    }
    else if (need == 0)
    {
      memcpy((m_vtlQueue->dataBuff + m_vtlQueue->front), src, len);
      m_vtlQueue->front = 0;
      m_vtlQueue->flag = 0;
      return len;
    }
    else
    {
      memcpy((m_vtlQueue->dataBuff + m_vtlQueue->front), src, len);
      m_vtlQueue->front = m_vtlQueue->front + len;
    }
    return len;
  }
  else
  {
    if (m_vtlQueue->rear - m_vtlQueue->front < len)
    {
      return -1;
    }
    else
    {
      memcpy((m_vtlQueue->dataBuff + m_vtlQueue->front), src, len);
      m_vtlQueue->front = m_vtlQueue->front + len;
      return len;
    }
  }
  return -1;
}

int CShareQueue::DeQue(void *dst, int len)
{
  if (m_vtlQueue->flag == 0)
  {
    int need = m_vtlQueue->rear + len - m_vtlQueue->buffSize;
    if (need > 0)
    {
      if (need > m_vtlQueue->front)
      {
        return -1;
      }
      int sgpt = len - need;
      memcpy(dst, (m_vtlQueue->dataBuff + m_vtlQueue->rear), sgpt);
      memcpy(((char *)dst + sgpt), m_vtlQueue->dataBuff, need);
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, sgpt);
      memset(m_vtlQueue->dataBuff, 0, need);
      m_vtlQueue->rear = need;
      m_vtlQueue->flag = 1;
      return len;
    }
    else if (need == 0)
    {
      memcpy((m_vtlQueue->dataBuff + m_vtlQueue->rear), dst, len);
      m_vtlQueue->rear = 0;
      m_vtlQueue->flag = 1;
      return len;
    }
    else
    {
      memcpy(dst, (m_vtlQueue->dataBuff + m_vtlQueue->rear), len);
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, len);
      m_vtlQueue->rear = m_vtlQueue->rear + len;
    }
    return len;
  }
  else
  {
    if (m_vtlQueue->front - m_vtlQueue->rear < len)
    {
      return -1;
    }
    else
    {
      memcpy(dst, (m_vtlQueue->dataBuff + m_vtlQueue->rear), len);
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, len);
      m_vtlQueue->rear = m_vtlQueue->rear + len;
      return len;
    }
  }
  return -1;
}

int CShareQueue::FtQue(void *dst, int len)
{
  if (m_vtlQueue->flag == 0)
  {
    int need = m_vtlQueue->rear + len - m_vtlQueue->buffSize;
    if (need > 0)
    {
      if (need > m_vtlQueue->front)
      {
        return -1;
      }
      int sgpt = len - need;
      memcpy(dst, (m_vtlQueue->dataBuff + m_vtlQueue->rear), sgpt);
      memcpy(((char *)dst + sgpt), m_vtlQueue->dataBuff, need);
      return len;
    }
    else
    {
      memcpy(dst, (m_vtlQueue->dataBuff + m_vtlQueue->rear), len);
    }
    return len;
  }
  else
  {
    if (m_vtlQueue->front - m_vtlQueue->rear < len)
    {
      return -1;
    }
    else
    {
      memcpy(dst, (m_vtlQueue->dataBuff + m_vtlQueue->rear), len);
      return len;
    }
  }
  return -1;
}

int CShareQueue::GetDataWithIdx(void *dst, int idxbegin, int len)
{
  if (m_vtlQueue->flag == 0)
  {
    int need = idxbegin + len - m_vtlQueue->buffSize;
    if (need > 0)
    {
      if (need > m_vtlQueue->front)
      {
        return -1;
      }
      int sgpt = len - need;
      memcpy(dst, (m_vtlQueue->dataBuff + idxbegin), sgpt);
      memcpy(((char *)dst + sgpt), m_vtlQueue->dataBuff, need);
      return len;
    }
    else
    {
      memcpy(dst, (m_vtlQueue->dataBuff + idxbegin), len);
    }
    return len;
  }
  else
  {
    if (m_vtlQueue->front - idxbegin < len)
    {
      return -1;
    }
    else
    {
      memcpy(dst, (m_vtlQueue->dataBuff + idxbegin), len);
      return len;
    }
  }
  return -1;
}

int CShareQueue::DlQue(int len)
{
  if (m_vtlQueue->flag == 0)
  {
    int need = m_vtlQueue->rear + len - m_vtlQueue->buffSize;
    if (need > 0)
    {
      if (need > m_vtlQueue->front)
      {
        return -1;
      }
      int sgpt = len - need;
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, sgpt);
      memset(m_vtlQueue->dataBuff, 0, need);
      m_vtlQueue->rear = need;
      m_vtlQueue->flag = 1;
      return len;
    }
    else if (need == 0)
    {
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, len);
      m_vtlQueue->rear = 0;
      m_vtlQueue->flag = 1;
      return len;
    }
    else
    {
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, len);
      m_vtlQueue->rear = m_vtlQueue->rear + len;
    }
    return len;
  }
  else
  {
    if (m_vtlQueue->front - m_vtlQueue->rear < len)
    {
      return -1;
    }
    else
    {
      memset((m_vtlQueue->dataBuff + m_vtlQueue->rear), 0, len);
      m_vtlQueue->rear = m_vtlQueue->rear + len;
      return len;
    }
  }
  return -1;
}

int CShareQueue::FindMsgFrame(int &idxbegin, int &idxend, int &len)
{
  int idx = m_vtlQueue->rear;
  idxbegin = idxend = -1;
  int sizeCount = 0;
  int queueSize = (m_vtlQueue->buffSize + m_vtlQueue->front - m_vtlQueue->rear) % m_vtlQueue->buffSize;
  if (queueSize == 0 && m_vtlQueue->flag == 0)
  {
    queueSize = m_vtlQueue->buffSize;
  }
  while (sizeCount < queueSize)
  {
    sizeCount++;
    // DEBUG_LOG("idx="<<idx<<",idxbegin="<<idxbegin<<",idxend="<<idxend);
    if (*(m_vtlQueue->dataBuff + idx) == 2 && idxbegin == -1)
    {
      idxbegin = idx;
    }
    else if (*(m_vtlQueue->dataBuff + idx) == 3 && idxbegin != -1)
    {
      idxend = idx + 1;
      len = (idxend + m_vtlQueue->buffSize - idxbegin) % m_vtlQueue->buffSize;
      break;
    }
    else
    {
    }
    idx = (idx + 1) % m_vtlQueue->buffSize;
  }
  DEBUG_LOG("idx=" << idx << ",idxbegin=" << idxbegin << ",idxend=" << idxend);
  if (idxbegin == -1 || idxend == -1)
  {
    ERROR_LOG("idxBegin=" << idxbegin << ",idxEnd=" << idxend);
    ERROR_LOG("Count find start_end flag pair!ret=" << ERRCODE_DATA_CONFUSION);
    return ERRCODE_DATA_CONFUSION;
  }
  return 0;
}

int CShareQueue::CheckMsgFrame(const int &len)
{
  MsgFrame msgFrame;
  GetDataWithIdx((char *)&msgFrame, m_vtlQueue->rear + 1, sizeof(MsgFrame));
  DEBUG_LOG("msgFrame.len=" << msgFrame.len << ",their=" << len - 2 - (int)sizeof(MsgFrame));
  if (msgFrame.len == len - 2 - (int)sizeof(MsgFrame))
  {
    return 0;
  }
  return ERRCODE_DATA_CONFUSION;
}

int CShareQueue::GetFreeSize()
{
  if (m_vtlQueue->rear == m_vtlQueue->front && m_vtlQueue->flag == 1)
  {
    return m_vtlQueue->buffSize;
  }
  return (m_vtlQueue->rear + m_vtlQueue->buffSize - m_vtlQueue->front) % m_vtlQueue->buffSize;
}

int CShareQueue::UndoQue(int len)
{
  return m_vtlQueue->front = (m_vtlQueue->front + m_vtlQueue->buffSize - len) % m_vtlQueue->buffSize;
}

int CShareQueue::Push(const void *src, int len)
{
  char head = 2;
  char tail = 3;
  int res = 0;
  m_rwMtx.Lock();
  if (GetFreeSize() < len + (int)sizeof(MsgFrame) + 2)
  {
    m_rwMtx.UnLock();
    return ERRCODE_PARAGRAM_FAILED;
  }
  do
  {
    /*head+msgFrame+data+end形式依次存入如果本次如队不成功，
    那么要删除之前的脏数据,并且返回错误*/
    if (EnQue(&head, 1) < 0)
    {
      res = ERRCODE_ILLEGAL_ADDR;
      break;
    }
    MsgFrame msgFrame;
    DEBUG_LOG("msgFrame.len=" << len);
    msgFrame.len = len;
    msgFrame.timeStamp = Milliseconds();
    if (EnQue(&msgFrame, sizeof(MsgFrame)) < 0)
    {
      res = ERRCODE_ILLEGAL_ADDR;
      sprintf(ShmErrorMsgBuff, "EnQue Error! ErrorCode=%d\n", ERRCODE_ILLEGAL_ADDR);
      UndoQue(1);
      break;
    }
    if (EnQue(src, len) < 0)
    {
      res = ERRCODE_ILLEGAL_ADDR;
      sprintf(ShmErrorMsgBuff, "EnQue Error! ErrorCode=%d\n", ERRCODE_ILLEGAL_ADDR);
      UndoQue(sizeof(MsgFrame) + 1);
      break;
    }
    if (EnQue(&tail, 1) < 0)
    {
      res = ERRCODE_ILLEGAL_ADDR;
      sprintf(ShmErrorMsgBuff, "EnQue Error! ErrorCode=%d\n", ERRCODE_ILLEGAL_ADDR);
      UndoQue(sizeof(MsgFrame) + 1 + len);
      break;
    }
    m_vtlQueue->msgNum++;
  } while (0);
  m_rwMtx.UnLock();
  return res;
}

// 返回值<1表示没有那么多的数据出队
int CShareQueue::Front(void *dst)
{
  return PopOut(dst, 0);
}

int CShareQueue::Pop()
{
  char dst[1] = {0};
  return PopOut(dst, 1);
}

/*出队函数flag默认为2
 *flag:0->只取头元素不出队；1->只出队不取元素；2->取完元素后出队
 *建议使用此函数出队，front和pop分开，同步过程中可能出现变化
 */
int CShareQueue::PopOut(void *dst, int flag)
{
  int ret = 0;
  int idxbegin, idxend, len;
  if (Empty())
  {
    ret = ERRCODE_QUE_EMPTY;
    sprintf(ShmErrorMsgBuff, "EnQue Empty! ErrorCode=%d\n", ERRCODE_QUE_EMPTY);
    return ERRCODE_QUE_EMPTY;
  }
  m_rwMtx.Lock();
  do
  {
    if (FindMsgFrame(idxbegin, idxend, len) < 0)
    {
      ret = ERRCODE_DATA_CONFUSION;
      sprintf(ShmErrorMsgBuff, "FindMsgFrame Empty! ErrorCode=%d\n", ERRCODE_DATA_CONFUSION);
      break;
    }
    m_vtlQueue->rear = idxbegin; // 校正一下队尾的位置
    ret = CheckMsgFrame(len);
    if (ret < 0)
    {
      break;
    }
    if (flag == 1)
    {
      // 将一个元素出队
      DlQue(len);
      m_vtlQueue->msgNum--;
      m_vtlQueue->lastPopTime = Milliseconds();
    }
    else
    {
      // 取元素中除去封装信息的数据
      GetDataWithIdx(dst, (idxbegin + sizeof(MsgFrame) + 1) % m_vtlQueue->buffSize, len - sizeof(MsgFrame) - 2);
      if (flag == 2)
      {
        // 将此元素出队
        DlQue(len);
        m_vtlQueue->msgNum--;
        m_vtlQueue->lastPopTime = Milliseconds();
      }
    }
  } while (0);
  m_rwMtx.UnLock();
  return ret;
}

int CShareQueue::Size()
{
  int res = 0;
  m_rwMtx.Lock();

  res = (m_vtlQueue->buffSize + m_vtlQueue->front - m_vtlQueue->rear) % m_vtlQueue->buffSize;
  // 如果是满的,queueSize = buffSize
  if (res == 0 && m_vtlQueue->flag == 0)
  {
    res = m_vtlQueue->buffSize;
  }
  m_rwMtx.UnLock();
  return res;
}

bool CShareQueue::Empty()
{
  // 如果当前队列中的长度小于消息帧头的长度，那么就算是空的
  if (Size() < (int)sizeof(MsgFrame))
  {
    return true;
  }
  return false;
}

bool CShareQueue::Full()
{
  // 如果当前队列的空余长度小于消息帧头的长度，那么就算是满的
  if (m_iMaxBuffSize - Size() < (int)sizeof(MsgFrame))
  {
    return true;
  }
  return false;
}
int CShareQueue::GetMsgNum()
{
  int num = 0;
  m_rwMtx.Lock();
  num = m_vtlQueue->msgNum;
  m_rwMtx.UnLock();
  return num;
}

long long CShareQueue::GetLastPopTime()
{
  long long lastTime = 0;
  m_rwMtx.Lock();
  lastTime = m_vtlQueue->lastPopTime;
  m_rwMtx.UnLock();
  return lastTime;
}

void CShareQueue::DeleteOuttimeMsg(long long referTime)
{
  int ret, idxbegin, idxend, len;
  ret = 0, len = 0;
  m_rwMtx.Lock();
  while (1)
  {
    if (FindMsgFrame(idxbegin, idxend, len) < 0)
    {
      ret = ERRCODE_DATA_CONFUSION;
      sprintf(ShmErrorMsgBuff, "FindMsgFrame Empty! ErrorCode=%d\n", ERRCODE_DATA_CONFUSION);
      break;
    }
    m_vtlQueue->front = idxbegin;
    ret = CheckMsgFrame(len);
    if (ret < 0)
    {
      sprintf(ShmErrorMsgBuff, "FindMsgFrame Empty! ErrorCode=%d\n", ERRCODE_DATA_CONFUSION);
      break;
    }
    MsgFrame msgFrame;
    GetDataWithIdx(&msgFrame, (m_vtlQueue->front + 1) % m_vtlQueue->buffSize, sizeof(MsgFrame));
    if (msgFrame.timeStamp < referTime)
    {
      DlQue(len);
      m_vtlQueue->msgNum--;
      m_vtlQueue->lastPopTime = Milliseconds();
    }
  }
  m_rwMtx.UnLock();
}

void CShareQueue::ShmPrint()
{
  m_rwMtx.Lock();
  DEBUGV_LOG("-----Status:front=" << m_vtlQueue->front << "; rear=" << m_vtlQueue->rear << ";flag=" << m_vtlQueue->flag << "----\n");
  DEBUGV_LOG("\n-------------------------------------------\n");
  for (int i = 0; i < m_vtlQueue->buffSize - 4; i = i + 4)
  {
    DEBUGV_LOG("0x" << hex << *(unsigned int *)(m_vtlQueue->dataBuff + i) << " ");
  }
  DEBUGV_LOG(dec << "\n--------------------------------------\n");
  m_rwMtx.UnLock();
}
