# Build stage
FROM golang:1.22.3-alpine3.19 AS builder 
# เติม AS ตามด้วยชื่อ stage นั้นเอง (จะทำเมื่อเราเขียน multi-stage นะ)
WORKDIR /app
COPY . .
RUN go build -o main main.go

# ทำ multi-stage build เพื่อลด size ของ image
# เหตุผลที่ size มันใหญ่เพราะว่ามันมี golang และ package ที่ต้องการใน project แต่สิ่งที่เราต้องการจริงๆเพื่อทำการรัน app นั้นมีแค่ binary file หลังจากรัน go build command ต่างหาก # เราต้องการแค่ binary file เท่านั้นแม้แต่ original golang code ก็ไม่จำเป็น # ซึ่งเราสามารถทำได้โดยทำเป็น multi-stage นั้นเอง

# Run stage
FROM alpine:3.19
# alpine:3.19 ใช้ alpine image เดียวกับที่ใช้ใน builder stage
WORKDIR /app
COPY --from=builder /app/main .
# copy มาแค่ file ใน /app/main เท่านั้น

# ทำให้จาก 588MB เมื่อ build ใหม่โดย multi-stage script นี้เหลือเป็น 21.2MB ทันที (ลดเยอะมาก)

EXPOSE 8080 
CMD [ "/app/main" ]
