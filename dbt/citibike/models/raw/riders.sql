{{ config(materialized='table') }}

select json from cdp_dev.raw.riders
