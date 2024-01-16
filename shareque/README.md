### 说明
用于说明共享内存的使用方法

### 构建方法
```
make
```
### 使用方法
1. 首先创建共享文件
```
touch /tmp/Shm
touch /tmp/SemMtx
chmod 777 /tmp/Shm
chmod 777 /tmp/SemMtx
```

2. 开两个终端都进入 bin 目录
一个终端执行
```
./WriteShm
``` 
另一个终端执行
```
./ReadShm
``` 

3. 查看日志对照远码也就清楚了共享内存的使用方法


