#!/bin/sh
# เนื่องจากคุณจะรันใน alpine image ทำให้มันจะรันใน /bin/sh shell (bash shell จะใช้ไม่ได้ใน image นี้)

# รัน "chmod +x start.sh" ที่ file นี้เพื่อให้สามารถ execute ได้ด้วยนะ

set -e

# ทำการรัน migrate up
echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up
# DB_SOURCE จาก app.env นั้นแหละ

# ทำการ start app
echo "start the app"
exec "$@" # $@ หมายความว่า เอาทุก parameters ใส่ลง script แล้วรันมัน