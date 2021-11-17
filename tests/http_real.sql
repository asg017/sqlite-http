.load dist/http0.so

.mode csv
.headers on
.bail on

select * from http_get('https://api.metro.net/agencies/lametro/routes/18/runs/');