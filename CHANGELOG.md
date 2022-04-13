# Changelog

## Version 5-5-2022
- [Release](https://github.com/amberflo/metering-go/commit/b26e75a17d40c32de0c81a5b1bf08caa7ad6b467)
    - Add CustomerEmail to Customer
    - Update CreateTime and UpdateTime to integer for Unix Epoch

## Version 5-21-2021
- [Release](https://github.com/amberflo/metering-go/commit/eec4346de7b775e5a98e4852759a2261ee90da44) 
    - Update MeterMessage definition: 
        - Rename MeterName to MeterApiName
        - Rename UtcTimeMillis to MeterTimeInMillis
    - Update Customer definition: 
        - Add UpdateTime, CreateTime properties


## Version 4-2-2021
- [Release](https://github.com/amberflo/metering-go/commit/76ce7735145812a3bb12bfb0e4b6e7c2b264cc0a) 
    - Published Amberflo GO SDK with following functionality:
    - Added metering client
    - Configure metering client: BatchSize, IntervalSeconds, Debug
    - Create customers in Amberflo: ```Metering.AddorUpdateCustomer(customer)```
    - Ingest meters in Amberflo ```Metering.AddorUpdateCustomer(customer)```
    - Usage client to query usage in Amberflo and get a JSON response ```UsageClient.GetUsage(UsagePayload)```






