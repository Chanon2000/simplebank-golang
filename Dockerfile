# Build stage
FROM golang:1.22.3-alpine3.19 AS builder 
# เติม AS ตามด้วยชื่อ stage นั้นเอง (จะทำเมื่อเราเขียน multi-stage นะ)
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl 
# เนื่องจากโดย default นั้น alpine image มันไม่มี curl command
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz
# ทำการ download migrate command

# ทำ multi-stage build เพื่อลด size ของ image
# เหตุผลที่ size มันใหญ่เพราะว่ามันมี golang และ package ที่ต้องการใน project แต่สิ่งที่เราต้องการจริงๆเพื่อทำการรัน app นั้นมีแค่ binary file หลังจากรัน go build command ต่างหาก # เราต้องการแค่ binary file เท่านั้นแม้แต่ original golang code ก็ไม่จำเป็น # ซึ่งเราสามารถทำได้โดยทำเป็น multi-stage นั้นเอง

# Run stage
FROM alpine:3.19
# alpine:3.19 ใช้ alpine image เดียวกับที่ใช้ใน builder stage
WORKDIR /app
COPY --from=builder /app/main .
# copy มาแค่ file ใน /app/main เท่านั้น

COPY --from=builder /app/migrate ./migrate
# ย้าย migrate command เช้าที่ /usr/bin/migrate # คล้ายกับที่ทำให้ ci.yml แต่เปลี่ยน directory

# ทำให้จาก 588MB เมื่อ build ใหม่โดย multi-stage script นี้เหลือเป็น 21.2MB ทันที (ลดเยอะมาก)
COPY app.env .
# เพื่อ copy app.env เข้า image # โดยจะเห็นว่าใน app.env มี TOKEN_SYMMETRIC_KEY เก็บอยู่ซึ่งถ้าใน production image นั้นเราไม่ควรเอามันไปเก็บไว้ ซึ่งเราจะเรียนวิธีแทน production config ใน lecture อื่น
COPY start.sh .
# เอา start.sh เข้า image
COPY wait-for.sh .
COPY db/migration ./migration
# ทำการ copy migration folder ไปไว้ใน image ด้วย

EXPOSE 8080 
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]
# ทำการรัน start.sh ที่ ENTRYPOINT ซึ่งเมื่อ CMD ใช้ร่วมกับ ENTRYPOINT มันจะทำตัวเปลี่ยนเหมือน parameters นั้นทำให้ /app/main ของ CMD นี้แหละที่จะ pass เข้าไปเป็น parameter ใน $@ ที่เขียนใน start.sh นั้นเอง
# รัน CMD กับ ENTRYPOINT ร่วมกันไม่ต้องจากรัน ENTRYPOINT [ "/app/start.sh", "/app/main" ] เลยนะ แต่เขียนแยกกันแบบข้างบนมันเพิ่มความ flexibility


# รัน docker-compose up เพื่อรันตามปกติได้เลย หรือ docker-compose up --build ไปเลยจะได้ไม่ต้อง up down บ่อยๆ