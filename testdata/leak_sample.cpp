// leak_sample.cpp - Sample file with intentional memory leaks

#include <iostream>

class LeakyClass {
private:
    int* buffer;
    char* name;
    double* data;

public:
    LeakyClass(int size) {
        buffer = new int[size];      // Line 13: Allocated
        name = new char[100];        // Line 14: Allocated
        data = new double(3.14);     // Line 15: Allocated
    }

    ~LeakyClass() {
        delete[] buffer;             // OK: buffer is deleted
        // name is NOT deleted - LEAK!
        // data is NOT deleted - LEAK!
    }
};

class ArrayMismatch {
private:
    int* arr;

public:
    ArrayMismatch() {
        arr = new int[50];           // Allocated as array
    }

    ~ArrayMismatch() {
        delete arr;                   // Wrong! Should be delete[]
    }
};

class ReassignmentLeak {
private:
    int* ptr;

public:
    ReassignmentLeak() {
        ptr = new int(10);
    }

    void reset() {
        ptr = new int(20);           // LEAK: old ptr not deleted before reassignment
    }

    ~ReassignmentLeak() {
        delete ptr;                   // Only deletes the latest allocation
    }
};

class NoDestructor {
private:
    int* leaked;

public:
    NoDestructor() {
        leaked = new int[100];       // Allocated but no destructor to clean up!
    }
    // No destructor - memory will leak!
};
