package dash

import (
   "encoding/xml"
   "fmt"
   "os"
   "testing"
)

func Test_Info(t *testing.T) {
   for name := range tests {
      reps, err := read_file(name)
      if err != nil {
         t.Fatal(err)
      }
      fmt.Println(name)
      for i, rep := range reps {
         if i >= 1 {
            fmt.Println()
         }
         fmt.Println(rep)
      }
      fmt.Println()
   }
}

var tests = map[string]string{
   "mpd/amc.mpd": "",
   "mpd/hulu.mpd": "content-0",
   "mpd/nbc.mpd": "",
   "mpd/paramount.mpd": "0",
   "mpd/roku.mpd": "1",
}

func read_file(name string) ([]*Representation, error) {
   text, err := os.ReadFile(name)
   if err != nil {
      return nil, err
   }
   var m Media
   if err := xml.Unmarshal(text, &m); err != nil {
      return nil, err
   }
   return m.Representation(tests[name])
}
