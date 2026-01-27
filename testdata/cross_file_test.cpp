// cross_file_test.cpp - Implementation for cross-file analysis test

#include "cross_file_test.h"

DataManager::DataManager() {
  buffer = new int[1024]; // Allocation in cpp
  name = new char[256];   // Another allocation
}

DataManager::~DataManager() {
  delete[] buffer; // Proper cleanup
                   // name is NOT deleted - LEAK!
}

void DataManager::processData() {
  // Process the buffer
}
