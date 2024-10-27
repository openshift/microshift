## Group Organization in Bootc Layer

The following guiding principles should be used when adding more artifacts
to each group. 

> Important: Keep balanced build times within each group and maximize caching
> of artifacts independent of the current source code.

|Group |Build Time|Cached|Description|
|------|----------|------|-----------|
|group0| Short    | Yes  | Basic prerequisites
|group1| Long     | Yes  | Artifacts independent of current sources
|group2| Average  | No   | Current source prerequisites
|group3| Average  | No   | Current source artifacts on RHEL  
|group4| Average  | No   | Current source artifacts on CentOS
