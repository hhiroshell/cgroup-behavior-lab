package com.example.resourcemonitor;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.text.DecimalFormat;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.concurrent.TimeUnit;

public class ResourceMonitorApp {

    private static final DecimalFormat DF = new DecimalFormat("#,###");
    private static final DecimalFormat PERCENT_DF = new DecimalFormat("#0.00");
    private static final DateTimeFormatter TIME_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    public static void main(String[] args) {
        System.out.println("=".repeat(80));
        System.out.println("Resource Monitor Started");
        System.out.println("=".repeat(80));

        // Determine cgroup version
        CgroupVersion cgroupVersion = detectCgroupVersion();
        System.out.println("Detected cgroup version: " + cgroupVersion);
        System.out.println("=".repeat(80));
        System.out.println();

        // Monitor resources continuously
        while (true) {
            try {
                printResourceInfo(cgroupVersion);
                TimeUnit.SECONDS.sleep(5);
            } catch (InterruptedException e) {
                System.err.println("Interrupted, exiting...");
                Thread.currentThread().interrupt();
                break;
            } catch (Exception e) {
                System.err.println("Error reading resources: " + e.getMessage());
                e.printStackTrace();
            }
        }
    }

    private static void printResourceInfo(CgroupVersion version) {
        String timestamp = LocalDateTime.now().format(TIME_FORMATTER);
        System.out.println("[" + timestamp + "]");
        System.out.println("-".repeat(80));

        // CPU Information
        printCpuInfo(version);
        System.out.println();

        // Memory Information
        printMemoryInfo(version);
        System.out.println();

        System.out.println("=".repeat(80));
        System.out.println();
    }

    private static void printCpuInfo(CgroupVersion version) {
        System.out.println("CPU Resources:");

        // Available processors to JVM
        int availableProcessors = Runtime.getRuntime().availableProcessors();
        System.out.println("  Available Processors (JVM): " + availableProcessors);

        // Read cgroup CPU quota and period
        try {
            if (version == CgroupVersion.V2) {
                Path cpuMaxPath = Paths.get("/sys/fs/cgroup/cpu.max");
                if (Files.exists(cpuMaxPath)) {
                    String content = Files.readString(cpuMaxPath).trim();
                    String[] parts = content.split("\\s+");
                    if (parts.length == 2) {
                        String quota = parts[0];
                        String period = parts[1];
                        System.out.println("  cgroup v2 cpu.max: " + quota + " " + period);

                        if (!quota.equals("max")) {
                            long quotaValue = Long.parseLong(quota);
                            long periodValue = Long.parseLong(period);
                            double cpuLimit = (double) quotaValue / periodValue;
                            System.out.println("  CPU Limit: " + PERCENT_DF.format(cpuLimit) + " cores");
                        } else {
                            System.out.println("  CPU Limit: unlimited");
                        }
                    }
                }

                // CPU usage
                Path cpuStatPath = Paths.get("/sys/fs/cgroup/cpu.stat");
                if (Files.exists(cpuStatPath)) {
                    String content = Files.readString(cpuStatPath);
                    System.out.println("  cgroup v2 cpu.stat:");
                    for (String line : content.split("\n")) {
                        if (!line.trim().isEmpty()) {
                            System.out.println("    " + line);
                        }
                    }
                }
            } else {
                // cgroup v1
                Path quotaPath = Paths.get("/sys/fs/cgroup/cpu/cpu.cfs_quota_us");
                Path periodPath = Paths.get("/sys/fs/cgroup/cpu/cpu.cfs_period_us");

                if (Files.exists(quotaPath) && Files.exists(periodPath)) {
                    long quota = Long.parseLong(Files.readString(quotaPath).trim());
                    long period = Long.parseLong(Files.readString(periodPath).trim());

                    System.out.println("  cgroup v1 cpu.cfs_quota_us: " + quota);
                    System.out.println("  cgroup v1 cpu.cfs_period_us: " + period);

                    if (quota > 0) {
                        double cpuLimit = (double) quota / period;
                        System.out.println("  CPU Limit: " + PERCENT_DF.format(cpuLimit) + " cores");
                    } else {
                        System.out.println("  CPU Limit: unlimited");
                    }
                }

                // CPU usage
                Path cpuacctUsagePath = Paths.get("/sys/fs/cgroup/cpu,cpuacct/cpuacct.usage");
                if (Files.exists(cpuacctUsagePath)) {
                    long usage = Long.parseLong(Files.readString(cpuacctUsagePath).trim());
                    System.out.println("  Total CPU usage (nanoseconds): " + DF.format(usage));
                }
            }
        } catch (IOException | NumberFormatException e) {
            System.out.println("  Unable to read cgroup CPU info: " + e.getMessage());
        }
    }

    private static void printMemoryInfo(CgroupVersion version) {
        System.out.println("Memory Resources:");

        Runtime runtime = Runtime.getRuntime();
        long maxMemory = runtime.maxMemory();
        long totalMemory = runtime.totalMemory();
        long freeMemory = runtime.freeMemory();
        long usedMemory = totalMemory - freeMemory;

        System.out.println("  JVM Max Memory: " + formatBytes(maxMemory));
        System.out.println("  JVM Total Memory: " + formatBytes(totalMemory));
        System.out.println("  JVM Used Memory: " + formatBytes(usedMemory));
        System.out.println("  JVM Free Memory: " + formatBytes(freeMemory));

        // Read cgroup memory limits
        try {
            if (version == CgroupVersion.V2) {
                Path memMaxPath = Paths.get("/sys/fs/cgroup/memory.max");
                if (Files.exists(memMaxPath)) {
                    String content = Files.readString(memMaxPath).trim();
                    if (content.equals("max")) {
                        System.out.println("  cgroup v2 memory.max: unlimited");
                    } else {
                        long limit = Long.parseLong(content);
                        System.out.println("  cgroup v2 memory.max: " + formatBytes(limit));
                    }
                }

                // Current memory usage
                Path memCurrentPath = Paths.get("/sys/fs/cgroup/memory.current");
                if (Files.exists(memCurrentPath)) {
                    long current = Long.parseLong(Files.readString(memCurrentPath).trim());
                    System.out.println("  cgroup v2 memory.current: " + formatBytes(current));
                }

                // Memory stats
                Path memStatPath = Paths.get("/sys/fs/cgroup/memory.stat");
                if (Files.exists(memStatPath)) {
                    String content = Files.readString(memStatPath);
                    System.out.println("  cgroup v2 memory.stat (selected):");
                    for (String line : content.split("\n")) {
                        if (line.startsWith("anon ") || line.startsWith("file ") ||
                            line.startsWith("kernel_stack ") || line.startsWith("slab ")) {
                            String[] parts = line.split("\\s+");
                            if (parts.length == 2) {
                                System.out.println("    " + parts[0] + ": " + formatBytes(Long.parseLong(parts[1])));
                            }
                        }
                    }
                }
            } else {
                // cgroup v1
                Path limitPath = Paths.get("/sys/fs/cgroup/memory/memory.limit_in_bytes");
                if (Files.exists(limitPath)) {
                    long limit = Long.parseLong(Files.readString(limitPath).trim());
                    // Very large values indicate no limit
                    if (limit > (1L << 60)) {
                        System.out.println("  cgroup v1 memory.limit_in_bytes: unlimited");
                    } else {
                        System.out.println("  cgroup v1 memory.limit_in_bytes: " + formatBytes(limit));
                    }
                }

                // Current memory usage
                Path usagePath = Paths.get("/sys/fs/cgroup/memory/memory.usage_in_bytes");
                if (Files.exists(usagePath)) {
                    long usage = Long.parseLong(Files.readString(usagePath).trim());
                    System.out.println("  cgroup v1 memory.usage_in_bytes: " + formatBytes(usage));
                }

                // Memory stats
                Path statPath = Paths.get("/sys/fs/cgroup/memory/memory.stat");
                if (Files.exists(statPath)) {
                    String content = Files.readString(statPath);
                    System.out.println("  cgroup v1 memory.stat (selected):");
                    for (String line : content.split("\n")) {
                        if (line.startsWith("cache ") || line.startsWith("rss ") ||
                            line.startsWith("mapped_file ") || line.startsWith("inactive_anon ")) {
                            String[] parts = line.split("\\s+");
                            if (parts.length == 2) {
                                System.out.println("    " + parts[0] + ": " + formatBytes(Long.parseLong(parts[1])));
                            }
                        }
                    }
                }
            }
        } catch (IOException | NumberFormatException e) {
            System.out.println("  Unable to read cgroup memory info: " + e.getMessage());
        }
    }

    private static CgroupVersion detectCgroupVersion() {
        // Check if cgroup v2 is mounted
        Path cgroupV2 = Paths.get("/sys/fs/cgroup/cgroup.controllers");
        if (Files.exists(cgroupV2)) {
            return CgroupVersion.V2;
        }
        return CgroupVersion.V1;
    }

    private static String formatBytes(long bytes) {
        if (bytes < 1024) {
            return bytes + " B";
        } else if (bytes < 1024 * 1024) {
            return PERCENT_DF.format(bytes / 1024.0) + " KB";
        } else if (bytes < 1024 * 1024 * 1024) {
            return PERCENT_DF.format(bytes / (1024.0 * 1024)) + " MB";
        } else {
            return PERCENT_DF.format(bytes / (1024.0 * 1024 * 1024)) + " GB";
        }
    }

    private enum CgroupVersion {
        V1, V2
    }
}
