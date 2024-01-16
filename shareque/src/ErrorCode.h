#ifndef __ERROR_CODE_H__
#define __ERROR_CODE_H__

typedef int ErrCode;

const ErrCode ERRCODE_SUCCESS = 0;
const ErrCode ERRCODE_SHAREQUE_BEGIN = -100;
const ErrCode ERRCODE_PARAGRAM_FAILED = ERRCODE_SHAREQUE_BEGIN - 1;  // 共享内存没有建立或者容量不够
const ErrCode ERRCODE_SHM_INIT_FAILED = ERRCODE_SHAREQUE_BEGIN - 2;  // 内存没有建立或者容量不够
const ErrCode ERRCODE_LOCK_INIT_FAILED = ERRCODE_SHAREQUE_BEGIN - 3; // 初始化内存锁失败
const ErrCode ERRCODE_QUE_INIT_FAILED = ERRCODE_SHAREQUE_BEGIN - 4;  // 初始化队列失败
const ErrCode ERRCODE_QUE_FULL = ERRCODE_SHAREQUE_BEGIN - 5;         // 队列已满
const ErrCode ERRCODE_QUE_EMPTY = ERRCODE_SHAREQUE_BEGIN - 6;        // 队列为空
const ErrCode ERRCODE_ILLEGAL_ADDR = ERRCODE_SHAREQUE_BEGIN - 7;     // 访问地址非法
const ErrCode ERRCODE_DATA_CONFUSION = ERRCODE_SHAREQUE_BEGIN - 8;   // 数据混乱

#endif // __ERROR_CODE_H__
