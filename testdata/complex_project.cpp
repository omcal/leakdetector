// complex_project.cpp - Simulates a real-world complex codebase

// =============================================================================
// DATABASE CONNECTION MANAGER
// =============================================================================
class DatabaseConnection {
private:
  char *connectionString;
  int *resultBuffer;
  char *queryCache;

public:
  DatabaseConnection(const char *connStr) {
    connectionString = new char[256];
    resultBuffer = new int[1024];
    queryCache = new char[4096];
  }

  void executeQuery() {
    // Some query logic
  }

  void clearCache() {
    delete[] queryCache;
    queryCache = new char[4096]; // Reassignment - previous properly deleted
  }

  ~DatabaseConnection() {
    delete[] connectionString;
    delete[] resultBuffer;
    // queryCache NOT deleted - LEAK!
  }
};

// =============================================================================
// GRAPHICS RENDERER
// =============================================================================
class Texture {
private:
  unsigned char *pixels;
  int *metadata;

public:
  Texture(int width, int height) {
    pixels = new unsigned char[width * height * 4];
    metadata = new int[10];
  }

  ~Texture() {
    delete[] pixels;
    delete[] metadata; // Properly cleaned
  }
};

class Renderer {
private:
  float *vertexBuffer;
  float *normalBuffer;
  int *indexBuffer;
  Texture *mainTexture; // Note: raw pointer to object

public:
  Renderer() {
    vertexBuffer = new float[10000];
    normalBuffer = new float[10000];
    indexBuffer = new int[5000];
    mainTexture = new Texture(1024, 1024);
  }

  void resize(int newSize) {
    // BUG: Not deleting old buffers before reassignment
    vertexBuffer = new float[newSize]; // LEAK - old not deleted
    normalBuffer = new float[newSize]; // LEAK - old not deleted
  }

  void cleanup() {
    delete[] vertexBuffer;
    delete[] normalBuffer;
    delete[] indexBuffer;
    delete mainTexture;
  }

  ~Renderer() {
    cleanup(); // Cleanup via method - should be detected
  }
};

// =============================================================================
// NETWORK PACKET HANDLER
// =============================================================================
class PacketHandler {
private:
  char *receiveBuffer;
  char *sendBuffer;
  int *packetIds;

public:
  PacketHandler() {
    receiveBuffer = new char[65536];
    sendBuffer = new char[65536];
    packetIds = new int[1000];
  }

  void processPacket() {
    char *tempBuffer = receiveBuffer; // Alias
                                      // Process using tempBuffer
  }

  ~PacketHandler() {
    delete[] receiveBuffer;
    delete[] sendBuffer;
    delete[] packetIds;
  }
};

// =============================================================================
// AUDIO ENGINE
// =============================================================================
class AudioBuffer {
private:
  float *samples;
  int *channelMap;

public:
  AudioBuffer(int numSamples) {
    samples = new float[numSamples];
    channelMap = new int[8];
  }

  // Missing destructor - LEAK!
};

class AudioMixer {
private:
  AudioBuffer *buffers[4]; // Array of pointers
  float *mixBuffer;

public:
  AudioMixer() {
    for (int i = 0; i < 4; i++) {
      buffers[i] = new AudioBuffer(44100);
    }
    mixBuffer = new float[44100];
  }

  ~AudioMixer() {
    for (int i = 0; i < 4; i++) {
      delete buffers[i];
    }
    delete[] mixBuffer;
  }
};

// =============================================================================
// MEMORY POOL (Complex cleanup pattern)
// =============================================================================
class MemoryPool {
private:
  char *pool;
  int *freeList;
  int *usedList;

public:
  MemoryPool(int size) {
    pool = new char[size];
    freeList = new int[size / 64];
    usedList = new int[size / 64];
  }

  void *allocate(int bytes) {
    return nullptr; // Simplified
  }

  void deallocateInternal() {
    delete[] pool;
    delete[] freeList;
  }

  void deallocateAll() {
    deallocateInternal(); // 2-level method chain
    delete[] usedList;
  }

  ~MemoryPool() { deallocateAll(); }
};

// =============================================================================
// FILE SYSTEM HANDLER
// =============================================================================
class FileHandle {
private:
  char *filename;
  char *buffer;
  int *permissions;

public:
  FileHandle(const char *name) {
    filename = new char[256];
    buffer = new char[8192];
    permissions = new int(0644);
  }

  void close() {
    delete[] filename;
    delete[] buffer;
    delete permissions;
  }

  ~FileHandle() { close(); }
};

// =============================================================================
// GAME ENTITY SYSTEM
// =============================================================================
class Component {
private:
  int *data;

public:
  Component() { data = new int[16]; }
  ~Component() { delete[] data; }
};

class Entity {
private:
  char *name;
  float *position;
  float *rotation;
  Component *components[10];
  int componentCount;

public:
  Entity(const char *entityName) {
    name = new char[64];
    position = new float[3];
    rotation = new float[4];
    componentCount = 0;
  }

  void addComponent() {
    if (componentCount < 10) {
      components[componentCount++] = new Component();
    }
  }

  void destroyComponents() {
    for (int i = 0; i < componentCount; i++) {
      delete components[i];
    }
  }

  ~Entity() {
    delete[] name;
    delete[] position;
    delete[] rotation;
    destroyComponents();
  }
};

// =============================================================================
// THREAD POOL (Deep nesting)
// =============================================================================
class ThreadPool {
private:
  int *threadIds;
  char *taskQueue;
  int *priorityQueue;

public:
  ThreadPool(int numThreads) {
    threadIds = new int[numThreads];
    taskQueue = new char[1024];
    priorityQueue = new int[256];
  }

  void releaseQueues() {
    delete[] taskQueue;
    delete[] priorityQueue;
  }

  void releaseThreads() { delete[] threadIds; }

  void releaseAll() {
    releaseQueues();
    releaseThreads();
  }

  void shutdown() { releaseAll(); }

  ~ThreadPool() {
    shutdown(); // 4-level deep: dtor -> shutdown -> releaseAll ->
                // releaseQueues/releaseThreads
  }
};

// =============================================================================
// PROBLEMATIC CLASS WITH MULTIPLE ISSUES
// =============================================================================
class ProblematicClass {
private:
  int *arr1;       // new[] but delete
  int *arr2;       // Missing delete
  int *single;     // Proper
  int *reassigned; // Reassignment leak

public:
  ProblematicClass() {
    arr1 = new int[100];
    arr2 = new int[200];
    single = new int(42);
    reassigned = new int(1);
  }

  void update() {
    reassigned = new int(2); // LEAK - old value not deleted
  }

  ~ProblematicClass() {
    delete arr1; // Wrong! Should be delete[]
    // arr2 not deleted - LEAK!
    delete single;     // OK
    delete reassigned; // OK (but the old one leaked in update())
  }
};
