# Cloud Data Platform

As enterprises adopt Cloud-native data warehouses, like Snowflake and BigQuery, complimentary tools and techniques are needed to develop data pipelines, establish sound data governance, and produce analytics at scale and reliably for their organizations.

This project was set up to demonstrate how we can use modern software engineering practices and tools to supplement a set of modern, open-source data tech stack to build and scale our Cloud-native data warehouse.

## Architecture

### Data Warehouse

While the goal of this project is to use open-source tools as much as possible, the de-facto data warehouse that most enterprises use today is Snowflake. So Snowflake is used our data warehouse for this project. We may add Druid as an sample data warehouse in later versions of this project.

### ELT

ETL, which stands for Extract, Transform, Load, is a technique that integrates data from one data source into a data warehouse. ETL was developed in the 1970's when compute and storage resources were scare - hence transform the datasets first and then load what is needed to the target data system. Because transformation is performed before data loading, data engineers would need to rebuild and rerun the entire pipeline.

With the advent of the cloud, we are no longer constrained by compute/storage resources and costs. In recent years, ELT, which stands for Extract, Load, Transform, emerged as a viable alternative to ETL for integrating data into a data warehouse from various data sources. We will be using ELT in this project.

## Tools

The following tools are used in this project:

| Component                       | Tool                                                                                | Version |
|---------------------------------|-------------------------------------------------------------------------------------|---------|
| Data warehouse                  | [Snowflake ](https://snowflake.com)                                                 | n/a     |
| Data extractor (EL)             | [Airbyte](https://github.com/airbytehq/airbyte)                                     | 0.43    |
| Transformaton tool (T)          | [Data Builder Tool (DBT)](https://github.com/dbt-labs/dbt-core)                     | 1.4     |
| Data unit test framework/runner | [Great Expectations (GX)](https://github.com/great-expectations/great_expectations) | 0.15    |
| Environment provisioning        | [Terraform](https://github.com/hashicorp/terraform)                                 | 1.3     |
| CI/CD                           | [Github Actions](https://github.com/features/actions)                               | n/a     |
 
### Snowflake

Here's a summary of Snowflake resources:

* Account
  * Storage resources (abstract term)
    * Databases - charged by storage
      * Schemas
        * Tables
        * Views
        * Objects
  * Compute resources (abstract term)
    * Warehouses - charged by size of the compute instance associated with a warehouse.
  * IAM (abstract term)
    * Users
    * Roles

We can combine different set of databases, warehouses, and users/roles whenever we provision a new Snowflake setup. Here's a great [article by dbt Labs on their recommended starter Snowflake setup](https://www.getdbt.com/blog/how-we-configure-snowflake/) that can scale up nicely in the future. Our terraform module adopts dbt Labs' pattern with 1 small modification: instead of 2 databases, we are using 1 database with 2 user-defined schemas.

## Setup

Let's start building...

### Provision Snowflake Resources

#### Snowflake Setup

1. If you haven't done so already, [sign up](https://signup.snowflake.com/) for a Snowflake account.

We need to set our Snowflake account for Terraform integration.

#### Snowflake Service Account for Terraform 

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
   >  We are granting `SYSADMIN` and `SECURITYADMIN` roles the service account out of convenience. Don't do this in a production environment.
   > 

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

#### Run Terraform

1. Run terraform to create a data warehouse and database on Snowflake.

   ```shell
   cd terraform
   source .env
   terraform init
   terraform plan
   terraform apply
   ```

## References

* Terraform
  * [Terraforming Snowflake](https://quickstarts.snowflake.com/guide/terraforming_snowflake/index.html)
  * [Terraform provider: Snowflake](https://github.com/Snowflake-Labs/terraform-provider-snowflake)
  * [Terraform Snowflake Provider Documentation](https://registry.terraform.io/providers/Snowflake-Labs/snowflake/latest/docs)
* Snowflake
  * [DBT Labs Recommended Snowflake Setup](https://www.getdbt.com/blog/how-we-configure-snowflake/)
