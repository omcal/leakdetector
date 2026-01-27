// clean_sample.cpp - Sample file with properly managed memory

#include <iostream>

class CleanClass {
private:
  int *buffer;
  char *name;

public:
  CleanClass(int size) {
    buffer = new int[size];
    name = new char[100];
  }

  ~CleanClass() {
    delete[] buffer;
    delete[] name;
  }
};

class CleanWithCleanup {
private:
  double *data;
  int *numbers;

public:
  CleanWithCleanup() {
    data = new double[10];
    numbers = new int[20];
  }

  void cleanup() {
    delete[] data;
    delete[] numbers;
    data = nullptr;
    numbers = nullptr;
  }

  ~CleanWithCleanup() {
    cleanup(); // Destructor calls cleanup method - this is OK
  }
};

class ProperReassignment {
private:
  int *ptr;

public:
  ProperReassignment() { ptr = new int(10); }

  void reset() {
    delete ptr; // Properly delete before reassignment
    ptr = new int(20);
  }

  ~ProperReassignment() { delete ptr; }
};
