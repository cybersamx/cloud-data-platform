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
 
## Setup

Let's start building...

### Snowflake Setup

1. If you haven't done so already, [sign up](https://signup.snowflake.com/) for a Snowflake account.


