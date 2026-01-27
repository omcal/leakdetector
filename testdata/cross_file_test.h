// cross_file_test.h - Header file for cross-file analysis test

#ifndef CROSS_FILE_TEST_H
#define CROSS_FILE_TEST_H

class DataManager {
private:
  int *buffer; // Pointer member declared in header
  char *name;  // Another pointer member

public:
  DataManager();  // Constructor declared
  ~DataManager(); // Destructor declared

  void processData();
};

#endif
