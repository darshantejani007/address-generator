# Regex based Wallet generator

### GoLang script to generate desired wallet addresses which match your regex pattern.

This is a brute force loop of running through random wallets and verifying if they match your patterns. It utilises multi-cores of your system to max out the CPU performance. 

### Setup

Create a `.env` file by copying the `.env.example` and modify it as per your system cores and requirements.
```bash
cp .env.example .env
```


Run the go project with following:
```bash
go get
go run .
```

The output is generated in the `wallets.txt`. As you can see from the `.env.example` file, there are 2 patterns as an input. 1st is `acceptable_pattern`, which is not your final desired input, but you may just want to print them out in case they appear. 2nd is the `final_pattern`, which if appeared, will halt the application and print it out as the final result. 