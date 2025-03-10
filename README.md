# paycor-driver-sync

This sample application synchronizes your drivers addresses from Paycor to the [Mike Albert Fleet Solutions API](https://developer.mikealbert.com/).


## Paycor Setup
You will need to have a Paycor account and have the Reporting API enabled. Create a report that returns the driver information you want to synchronize, in CSV format and include the following columns, in this order:
 1. Employee Number
 2. Last Name
 3. First Name
 4. Address 1
 5. Address 2
 6. City
 7. State/Province
 8. ZIP/Postal Code

## Configuration
The configuration file should be in YAML format and include the following information:

```yaml
paycor:
  host: secure.paycor.com
  publickey: Provided by Paycor
  privatekey: Provided by Paycor
  homeaddressesreport: Name of Paycor report that returns the driver information to process
mikealbert:
  clientid: Provided by Mike Albert
  clientsecret: Provided by Mike Albert
  endpoint: https://api.mikealbert.com/api/v1/
```

## Running the Application
The application can be run from the command line with the following command:

```bash
paycor-driver-sync -config paycor-driver-sync.yaml
```

This could also be run as a cron job or a scheduled task.