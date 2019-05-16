package lib

import (
	"testing"
)

func TestSingleAppend(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf(err.Error())
	}
	gsn, err := client.Append("Hello, World!")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if gsn < 1 {
		t.Fatalf("Record assigned invalid global sequence number %d", gsn)
	}
}

// func TestMultipleAppend(t *testing.T) {
// 	client, err := NewClient()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// 	set := set64.NewSet64()
// 	for i := 1; i <= 5; i++ {
// 		gsn, err := client.Append(fmt.Sprintf("Appending %d", i))
// 		if err != nil {
// 			t.Fatalf(err.Error())
// 		}
// 		if gsn < 1 {
// 			t.Fatalf("Record assigned invalid global sequence number %d", gsn)
// 		}
// 		if set.Contains(int64(gsn)) {
// 			t.Fatalf("Record assigned previously assigned global sequence number %d", gsn)
// 		}
// 		set.Add(int64(gsn))
// 	}
// }

// func TestMultipleClientsAppend(t *testing.T) {
// 	set := set64.NewSet64()
// 	for i := 1; i <= 5; i++ {
// 		client, err := NewClient()
// 		if err != nil {
// 			t.Fatalf(err.Error())
// 		}
// 		gsn, err := client.Append(fmt.Sprintf("Appending %d", i))
// 		if err != nil {
// 			t.Fatalf(err.Error())
// 		}
// 		if gsn < 1 {
// 			t.Fatalf("Record assigned invalid global sequence number %d", gsn)
// 		}
// 		if set.Contains(int64(gsn)) {
// 			t.Fatalf("Record assigned previously assigned global sequence number %d", gsn)
// 		}
// 		set.Add(int64(gsn))
// 	}
// }

// func TestSimpleSubscribe(t *testing.T) {
// 	client, err := NewClient()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// 	c, err := client.Subscribe(1)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}
// 	_, err = client.Append("Hello, World!")
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}
// 	resp := <-c
// 	if resp.Gsn < 1 {
// 		t.Fatalf("Record assigned invalid global sequence number %d", resp.Gsn)
// 	}
// }

// func TestMultipleSubscribe(t *testing.T) {
// 	client, err := NewClient()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
// 	c, err := client.Subscribe(1)
// 	if err != nil {
// 		t.Errorf(err.Error())
// 	}
// 	set := set64.NewSet64()
// 	for i := 1; i <= 5; i++ {
// 		_, err := client.Append(fmt.Sprintf("Appending %d", i))
// 		if err != nil {
// 			t.Fatalf(err.Error())
// 		}
// 		resp := <-c
// 		if resp.Gsn < 1 {
// 			t.Fatalf("Record assigned invalid global sequence number %d", resp.Gsn)
// 		}
// 		if set.Contains(int64(resp.Gsn)) {
// 			t.Fatalf("Duplicate subscribe responses received for global sequence number %d", resp.Gsn)
// 		}
// 		set.Add(int64(resp.Gsn))
// 	}
// }
