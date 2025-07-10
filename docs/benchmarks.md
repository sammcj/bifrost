# 📊 Bifrost Benchmarks

Bifrost has been tested under high load conditions to ensure optimal performance. The following results were obtained from benchmark tests running at 5000 requests per second (RPS) on different AWS EC2 instances.

---

## 🧪 Test Environment

### **1. t3.medium (2 vCPUs, 4GB RAM)**

- Buffer Size: 15,000
- Initial Pool Size: 10,000

### **2. t3.xlarge (4 vCPUs, 16GB RAM)**

- Buffer Size: 20,000
- Initial Pool Size: 15,000

---

## 📈 Performance Metrics

| Metric                    | t3.medium     | t3.xlarge      |
| ------------------------- | ------------- | -------------- |
| Success Rate              | 100.00%       | 100.00%        |
| Average Request Size      | 0.13 KB       | 0.13 KB        |
| **Average Response Size** | **`1.37 KB`** | **`10.32 KB`** |
| Average Latency           | 2.12s         | 1.61s          |
| Peak Memory Usage         | 1312.79 MB    | 3340.44 MB     |
| Queue Wait Time           | 47.13 µs      | 1.67 µs        |
| Key Selection Time        | 16 ns         | 10 ns          |
| Message Formatting        | 2.19 µs       | 2.11 µs        |
| Params Preparation        | 436 ns        | 417 ns         |
| Request Body Preparation  | 2.65 µs       | 2.36 µs        |
| JSON Marshaling           | 63.47 µs      | 26.80 µs       |
| Request Setup             | 6.59 µs       | 7.17 µs        |
| HTTP Request              | 1.56s         | 1.50s          |
| Error Handling            | 189 ns        | 162 ns         |
| Response Parsing          | 11.30 ms      | 2.11 ms        |
| **Bifrost's Overhead**    | **`59 µs\*`** | **`11 µs\*`**  |

_\*Bifrost's overhead is measured at 59 µs on t3.medium and 11 µs on t3.xlarge, excluding the time taken for JSON marshalling and the HTTP call to the LLM, both of which are required in any custom implementation._

**Note**: On the t3.xlarge, we tested with significantly larger response payloads (~10 KB average vs ~1 KB on t3.medium). Even so, response parsing time dropped dramatically thanks to better CPU throughput and Bifrost's optimized memory reuse.

**Disclaimer**: These metrics are measured without the UI enabled. When using the UI, there is no drop in performance - only memory usage increases due to the additional UI build being served.

---

## 🎯 Key Performance Highlights

- **Perfect Success Rate**: 100% request success rate under high load on both instances
- **Total Overhead**: Less than only _15µs added per request_ on average
- **Efficient Queue Management**: Minimal queue wait time (1.67 µs on t3.xlarge)
- **Fast Key Selection**: Near-instantaneous key selection (10 ns on t3.xlarge)
- **Improved Performance on t3.xlarge**:
  - 24% faster average latency
  - 81% faster response parsing
  - 58% faster JSON marshaling
  - Significantly reduced queue wait times

---

## ⚙️ Configuration Flexibility

One of Bifrost's key strengths is its flexibility in configuration. You can freely decide the tradeoff between memory usage and processing speed by adjusting Bifrost's configurations. This flexibility allows you to optimize Bifrost for your specific use case, whether you prioritize speed, memory efficiency, or a balance between the two.

- Higher buffer and pool sizes (like in t3.xlarge) improve speed but use more memory
- Lower configurations (like in t3.medium) use less memory but may have slightly higher latencies
- You can fine-tune these parameters based on your specific needs and available resources

### **Key Configuration Parameters**

- **Initial Pool Size**: Determines the initial allocation of resources
- **Buffer and Concurrency Settings**: Controls the queue size and maximum number of concurrent requests (adjustable per provider)
- **Retry and Timeout Configurations**: Customizable based on your requirements for each provider

---

## 🚀 Run Your Own Benchmarks

Curious? Run your own benchmarks. The [Bifrost Benchmarking](https://github.com/maximhq/bifrost-benchmarking) repo has everything you need to test it in your own environment.

---

## 🔗 Related Documentation

**🏛️ Curious how we handle scales of 10k+ RPS?** Check out our [System Architecture Documentation](./architecture/system-overview.md) for detailed insights into Bifrost's high-performance design, memory management, and scaling strategies.

- **[🌐 System Overview](./architecture/system-overview.md)** - High-level architecture components
- **[🔄 Request Flow](./architecture/request-flow.md)** - Request processing pipeline
- **[⚙️ Concurrency Model](./architecture/concurrency.md)** - Worker pools and threading details
- **[💡 Design Decisions](./architecture/design-decisions.md)** - Performance-related architectural choices
