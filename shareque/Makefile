CXX=g++
OBJ_DIR=./obj
BIN_DIR=./bin
SRC_DIR=./src
CFLAGS = -g -Wall -I$(SRC_DIR)

all:$(BIN_DIR)/ReadShm $(BIN_DIR)/WriteShm
$(OBJ_DIR)/ReadShm.o :
	$(CXX) $(CFLAGS) -c -o $@ $(SRC_DIR)/ReadShm.cpp
$(OBJ_DIR)/WriteShm.o :
	$(CXX) $(CFLAGS) -c -o $@ $(SRC_DIR)/WriteShm.cpp
$(OBJ_DIR)/ShareQueue.o :
	$(CXX) $(CFLAGS) -c -o $@ $(SRC_DIR)/ShareQueue.cpp
$(OBJ_DIR)/Semaphore.o :
	$(CXX) $(CFLAGS) -c -o $@ $(SRC_DIR)/Semaphore.cpp

$(BIN_DIR)/ReadShm:$(OBJ_DIR)/ReadShm.o $(OBJ_DIR)/ShareQueue.o $(OBJ_DIR)/Semaphore.o
	$(CXX) -o $@ $(OBJ_DIR)/ReadShm.o $(OBJ_DIR)/ShareQueue.o $(OBJ_DIR)/Semaphore.o 
$(BIN_DIR)/WriteShm:$(OBJ_DIR)/WriteShm.o $(OBJ_DIR)/ShareQueue.o $(OBJ_DIR)/Semaphore.o
	$(CXX) -o $@ $(OBJ_DIR)/WriteShm.o $(OBJ_DIR)/ShareQueue.o $(OBJ_DIR)/Semaphore.o 

clean:
	rm -rf $(BIN_DIR)/ReadShm $(BIN_DIR)/WriteShm $(OBJ_DIR)/*.o
