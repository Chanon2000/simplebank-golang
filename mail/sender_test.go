package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/chanon2000/simplebank/util"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() { // check ว่า ถ้ามี Short flag (ซึ่งถ้ามี Short() ก็จะเป็น true) ให้ทำการ skip test นี้
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="http://mick.guru">Tech School</a></p>
	`
	to := []string{"maymic2543@gmail.com"}
	attachFiles := []string{"../README.md"} // เอา README.md เป็น file ที่จะส่งไป

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles) // ทดสอบการส่ง mail จริงๆ // ซึ่งเมื่อทำการกดรัน test ก็จะเห็น mail ส่งไปที่ maymic2543@gmail.com จริงๆ นั้นเอง
	require.NoError(t, err) // ถ้าผ่าน ก็แสดงว่า SendEmail ของเราทำงานได้ตามปกติ
}
