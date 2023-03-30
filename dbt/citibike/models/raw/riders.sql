{{ config(materialized='table') }}

select val from cdp_dev.raw.riders
