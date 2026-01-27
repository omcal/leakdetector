// multi_level_test.cpp - Test for multi-level method call tracking

class MultiLevel {
private:
  int *data;

public:
  MultiLevel() { data = new int[100]; }

  void innerCleanup() {
    delete[] data; // Actual delete is 2 levels deep
  }

  void outerCleanup() {
    innerCleanup(); // Calls inner
  }

  ~MultiLevel() {
    outerCleanup(); // Destructor -> outer -> inner -> delete
  }
};

// This should be CLEAN - multi-level tracking should find the delete
