// alias_test.cpp - Test for pointer aliasing detection

class AliasLeak {
private:
  int *original;

public:
  AliasLeak() { original = new int(42); }

  void badAlias() {
    int *alias = original; // Pointer alias
    delete alias;          // Delete through alias
    delete original;       // Double delete! Should be flagged
  }

  ~AliasLeak() { delete original; }
};

class AliasSafe {
private:
  int *ptr;

public:
  AliasSafe() { ptr = new int(10); }

  void useAlias() {
    int *temp = ptr; // Create alias
    *temp = 20;      // Use it
                     // Don't delete temp - ptr will be deleted in dtor
  }

  ~AliasSafe() {
    delete ptr; // Clean - only one delete
  }
};
