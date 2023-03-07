# Simulator

Build the project for running a "simulated" transactional data source, from which Airbyte will extract.

## Overview

Snowflake offers a beginner's workshop called [Virtual Zero-to-Snowflake](https://s3.amazonaws.com/snowflake-workshop-lab/OnlineZTS_LabGuide.pdf) hands-on lab. The lab uses the sample data from Citibike, which all the data for past 7 years or so are available on an AWS S3 bucket for loading to a Snowflake database instance directly using a Snowflake feature called **Stage**.

But the goal of this project is to play around with Airbyte and its change data capture capabilities, so we want to be able to do the following:

* Create a data source containing the application transactional data, from where Airbyte will extract. The data source is a Postgres database.
* Application transactions (ie. trips) will be written to the data. A simulator written in Go will write to Postgres database over a period of time. This should allow Airbyte to extract the newer data that are introduced since the last data sync.

This directory contains the Docker Compose file for running postgres and the Go program for simulating incremental transactional data updates to the database.

## Setup

1. Open a shell and start Postgres by running:

   ```shell
   make docker-up
   ```

1. Build and run the simulator:

   ```shell
   make build
   bin/simulator start
   ```

1. When we are done, run the following to teardown and close out the resources:

   ```shell
   make docker-down
   ```

> **Note**
> 
> This project is still a work-in-progress. So the setup is very basic. I hope to improve the project further.
