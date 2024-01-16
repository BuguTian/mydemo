#ifndef __LOG_H__
#define __LOG_H__

#include <iostream>
#include <ctime>
#include <string>
#include <cstdio>
#include <sys/time.h>
#include <time.h>

using namespace std;

// 日志开关
#define TIME_LOG
#define DEBUG
#define INFO
#define ERROR
#define DEBUGV

// 日志主体
#ifndef TIME_LOG
#define TIME_PRINT()
#else
#define TIME_PRINT() time_print()
#endif

#ifndef DEBUGV
#define DEBUGV_LOG(log)
#else
#define DEBUGV_LOG(log) cout << log
#endif

#ifndef DEBUG
#define DEBUG_LOG(log)
#else
#define DEBUG_LOG(log) cout << TIME_PRINT() << "DEBUG " << log << "[" << __FILE__ << ":" << __LINE__ << "]" << endl
#endif

#ifndef INFO
#define INFO_LOG(log)
#else
#define INFO_LOG(log) cout << TIME_PRINT() << "INFO " << log << "[" << __FILE__ << ":" << __LINE__ << "]" << endl
#endif

#ifndef ERROR
#define ERROR_LOG(log)
#else
#define ERROR_LOG(log) cout << TIME_PRINT() << "ERROR " << log << "[" << __FILE__ << ":" << __LINE__ << "]" << endl
#endif

// 打印日志的时候可能需要打印时间
std::string inline time_print()
{
  time_t now_time;
  now_time = time(NULL);
  tm *t = localtime(&now_time);
  char time_s[50];
  sprintf(time_s, "%d-%02d-%02d %02d:%02d:%02d ",
          t->tm_year + 1900,
          t->tm_mon + 1,
          t->tm_mday,
          t->tm_hour,
          t->tm_min,
          t->tm_sec);
  return time_s;
}

// 取毫秒级别的时间戳
long long inline Milliseconds()
{
  struct timeval t;
  struct timezone tz;
  gettimeofday(&t, &tz);
  return (t.tv_sec * 1000 + t.tv_usec / 1000);
}
// 取微秒级别的时间戳
long long inline Microseconds()
{
  struct timeval t;
  struct timezone tz;
  gettimeofday(&t, &tz);
  return (t.tv_sec * 1000 * 1000 + t.tv_usec);
}

#endif // __LOG_H__
