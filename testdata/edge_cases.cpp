// edge_cases.cpp - Comprehensive edge case tests

// =============================================================================
// CASE 1: Basic missing delete (should detect)
// =============================================================================
class BasicLeak {
private:
  int *ptr;

public:
  BasicLeak() { ptr = new int(42); }
  ~BasicLeak() { /* missing delete */ }
};

// =============================================================================
// CASE 2: Proper cleanup (should pass - no leak)
// =============================================================================
class ProperCleanup {
private:
  int *data;
  char *buffer;

public:
  ProperCleanup() {
    data = new int[100];
    buffer = new char[256];
  }
  ~ProperCleanup() {
    delete[] data;
    delete[] buffer;
  }
};

// =============================================================================
// CASE 3: Array mismatch - new[] with delete (should detect ERROR)
// =============================================================================
class ArrayMismatchNewArray {
private:
  int *arr;

public:
  ArrayMismatchNewArray() { arr = new int[50]; }
  ~ArrayMismatchNewArray() { delete arr; } // Wrong! Should be delete[]
};

// =============================================================================
// CASE 4: Array mismatch - new with delete[] (should detect WARNING)
// =============================================================================
class ArrayMismatchNewSingle {
private:
  int *single;

public:
  ArrayMismatchNewSingle() { single = new int(10); }
  ~ArrayMismatchNewSingle() { delete[] single; } // Wrong! Should be delete
};

// =============================================================================
// CASE 5: Multiple pointers - partial cleanup (should detect 1 leak)
// =============================================================================
class PartialCleanup {
private:
  int *a;
  int *b;
  int *c;

public:
  PartialCleanup() {
    a = new int(1);
    b = new int(2);
    c = new int(3);
  }
  ~PartialCleanup() {
    delete a;
    delete c;
    // b is NOT deleted - LEAK!
  }
};

// =============================================================================
// CASE 6: Cleanup via method (multi-level tracking - should pass)
// =============================================================================
class CleanupViaMethod {
private:
  double *values;

public:
  CleanupViaMethod() { values = new double[1000]; }
  void cleanup() { delete[] values; }
  ~CleanupViaMethod() { cleanup(); }
};

// =============================================================================
// CASE 7: Deep nested cleanup - 3 levels (should pass)
// =============================================================================
class DeepNestedCleanup {
private:
  float *matrix;

public:
  DeepNestedCleanup() { matrix = new float[64]; }
  void level3() { delete[] matrix; }
  void level2() { level3(); }
  void level1() { level2(); }
  ~DeepNestedCleanup() { level1(); }
};

// =============================================================================
// CASE 8: Reassignment leak (should detect WARNING)
// =============================================================================
class ReassignmentLeak {
private:
  int *ptr;

public:
  ReassignmentLeak() { ptr = new int(1); }
  void reassign() {
    ptr = new int(2); // LEAK - old ptr not deleted
  }
  ~ReassignmentLeak() { delete ptr; }
};

// =============================================================================
// CASE 9: Proper reassignment (should pass)
// =============================================================================
class ProperReassignment {
private:
  int *ptr;

public:
  ProperReassignment() { ptr = new int(1); }
  void reassign() {
    delete ptr; // Properly delete first
    ptr = new int(2);
  }
  ~ProperReassignment() { delete ptr; }
};

// =============================================================================
// CASE 10: No destructor (should detect ERROR)
// =============================================================================
class NoDestructor {
private:
  int *leaked;

public:
  NoDestructor() { leaked = new int[100]; }
  // No destructor at all!
};

// =============================================================================
// CASE 11: Pointer alias - double delete (should detect ERROR)
// =============================================================================
class DoubleDeleteViaAlias {
private:
  int *original;

public:
  DoubleDeleteViaAlias() { original = new int(42); }
  void badFunction() {
    int *alias = original;
    delete alias;
    delete original; // Double delete!
  }
  ~DoubleDeleteViaAlias() { delete original; }
};

// =============================================================================
// CASE 12: Safe alias usage (should pass)
// =============================================================================
class SafeAliasUsage {
private:
  int *ptr;

public:
  SafeAliasUsage() { ptr = new int(10); }
  void useAlias() {
    int *temp = ptr;
    *temp = 20; // Just using it, not deleting
  }
  ~SafeAliasUsage() { delete ptr; }
};

// =============================================================================
// CASE 13: this-> prefix (should work same as without)
// =============================================================================
class ThisPointerStyle {
private:
  int *member;

public:
  ThisPointerStyle() { this->member = new int(5); }
  ~ThisPointerStyle() { delete this->member; }
};

// =============================================================================
// CASE 14: Multiple allocations same line (edge case)
// =============================================================================
class MultiplePointers {
private:
  int *x;
  int *y;

public:
  MultiplePointers() {
    x = new int(1);
    y = new int(2);
  }
  ~MultiplePointers() {
    delete x;
    delete y;
  }
};

// =============================================================================
// CASE 15: Virtual destructor (should work)
// =============================================================================
class VirtualDestructorBase {
protected:
  int *base_ptr;

public:
  VirtualDestructorBase() { base_ptr = new int(100); }
  virtual ~VirtualDestructorBase() { delete base_ptr; }
};

// =============================================================================
// CASE 16: Struct instead of class (should work same)
// =============================================================================
struct StructTest {
  int *data;
  StructTest() { data = new int[10]; }
  ~StructTest() { delete[] data; }
};

// =============================================================================
// CASE 17: Private inheritance cleanup
// =============================================================================
class InheritanceTest : private VirtualDestructorBase {
private:
  char *child_ptr;

public:
  InheritanceTest() { child_ptr = new char[50]; }
  ~InheritanceTest() { delete[] child_ptr; }
};
