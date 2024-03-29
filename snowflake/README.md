# Snowflake Provisioning using Terraform

This directory contains Terraform modules that can be execute to provision the Snowflake resources needed for this project.

The terraform project for provisioning the resources we need on Snowflake is configured for a single-tenant environment. This means that only one person can provision the resources using Terraform as the project only saves the terraform state locally and won't be persisting the state in the Cloud.

Here's a summary of Snowflake resources:

* Account
  * Storage resources
    * Databases - charged by storage
      * Schemas
        * Tables
        * Views
        * Objects
  * Compute resources
    * Warehouses - charged by size of the compute instance associated with a warehouse.
  * IAM resources
    * Users
    * Roles

We can combine different set of databases, warehouses, and users/roles whenever we provision a new Snowflake setup. Our terraform module is based on DBT Lab's [article by dbt Labs on their recommended starter Snowflake setup](https://www.getdbt.com/blog/how-we-configure-snowflake/), which allows a clean separation of environments and scaling.

## Setup

### Snowflake Account Setup

1. If you haven't done so already, [sign up](https://signup.snowflake.com/) for a Snowflake account.

We need to set our Snowflake account for Terraform integration.

### Snowflake Service Account for Terraform Setup

The instructions were adopted from Snowflake document entitled [Terraform Snowflake](https://quickstarts.snowflake.com/guide/terraforming_snowflake/index.html).

1. Create RSA keypair for the terraform service account (sa). For this sa, we will be using private/public key for authentication.

   ```shell
   ssh-keygen -t rsa -b 4096 -C 'snowflake sa key for terraform' -f ~/.ssh/snow-tf-sa
   
   openssl genrsa 4096 | openssl pkcs8 -topk8 -inform PEM -out ~/.ssh/snow-tf-sa -nocrypt
   openssl rsa -in ~/.ssh/snow-tf-sa -pubout -out ~/.ssh/snow-tf-sa.pub
   ```

1. Navigate to the admin console of your Snowflake account. The url should be `<account-locator>.<cloud-provider-region>.<cloud-provider>.snowflakecomputing.com/console/login`.
1. Sign in and select `ACCOUNTADMIN` role.
1. Create the service account `terraform` in Snowflake by running this SQL command - no database selected:

   ```snowflake
   -- Replace RSA_PUBLIC_KEY_HERE with an actual public key omitting -----BEGIN PUBLIC KEY----- and -----END PUBLIC KEY-----
   CREATE USER "terraform" RSA_PUBLIC_KEY='RSA_PUBLIC_KEY_HERE' DEFAULT_ROLE=PUBLIC MUST_CHANGE_PASSWORD=FALSE;
   GRANT ROLE SYSADMIN TO USER "terraform";
   GRANT ROLE SECURITYADMIN TO USER "terraform";
   ```

   > **Notes**
   >
   >  We are granting `SYSADMIN` and `SECURITYADMIN` roles to the service account out of convenience. Don't do this in a production environment.
   
1. Get the account locator and account cloud provider/region from the url of your console. There are 2 naming schemes:

  * `https://<account-locator>.<cloud-provider-region>.<cloud-provider>.snowflakecomputing.com`
  * `https://app.snowflake.com/<cloud-provider-region>.<cloud-provider>/<account-locator>`

1. Clone the `.env-example`, update the `.env` file with the information we received from the previous step and then source the file by running the following:

   ```shell
   cd terraform
   # Copy .env-example to .env (note: .env is ignored and won't be checked into git).
   cp .env-example .env
   vi .env
   source .env
   printenv | grep SNOWFLAKE
   ```

### Run Terraform

1. Run terraform to create a data warehouse and database on Snowflake.

   ```shell
   cd terraform
   source .env
   terraform init
   terraform plan
   terraform apply
   ```

## Notes

* Terraform module [snowflake_table](https://registry.terraform.io/providers/Snowflake-Labs/snowflake/latest/docs/resources/table) creates table columns on Snowflake that has double quotes "" wrapped around the column name, which translates to the following SQL command.

  ````snowflake
  create or replace TABLE CDP_DEV.RAW.TRIPS (
	  "GENDER" VARCHAR(16777216),
	  "BIKE_ID" VARCHAR(16777216)
  );
  ````

  Rather than the following:

  ```snowflake
  create or replace TABLE CDP_DEV.RAW.TRIPS (
	  GENDER VARCHAR(16777216),
	  BIKE_ID VARCHAR(16777216)
  );
  ```
  
  The way the terraform module creates the columns means that we must include the quotes in any references to the columns. So for now, we aren't using terraform to create tables. Instead that responsibility will be assigned to airbyte (for raw tables) and dbt (for transformed tables via dbt models).

## Future Work

* For convenience, we grant `SYSADMIN` to users `DBT` and `AIRBYTE` as our Airbyte operations need elevated privileges above the standard `USAGE` and `MODIFY`. We should not use `SYSADMIN` and assign specific privileges to the users instead.

## References

* [Terraforming Snowflake](https://quickstarts.snowflake.com/guide/terraforming_snowflake/index.html)
* [Terraform provider: Snowflake](https://github.com/Snowflake-Labs/terraform-provider-snowflake)
* [Terraform Snowflake Provider Documentation](https://registry.terraform.io/providers/Snowflake-Labs/snowflake/latest/docs)
* [Snowflake default password policy](https://docs.snowflake.com/en/user-guide/admin-user-management)

