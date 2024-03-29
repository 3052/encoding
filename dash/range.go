package dash

import "fmt"

type Representation struct {
   adaptation_set *adaptation_set
   Bandwidth int64 `xml:"bandwidth,attr"`
   BaseURL string
   Codecs string `xml:"codecs,attr"`
   Height int64 `xml:"height,attr"`
   ID string `xml:"id,attr"`
   MimeType string `xml:"mimeType,attr"`
   SegmentBase *struct {
      Initialization struct {
         Range Range `xml:"range,attr"`
      }
      IndexRange Range `xml:"indexRange,attr"`
   }
   SegmentTemplate *SegmentTemplate
   Width int64 `xml:"width,attr"`
}

type Range string

// range-start and range-end can both exceed 32 bits, so we must use 64 bit
func (r Range) Scan() (uint64, uint64, error) {
   var start, end uint64
   _, err := fmt.Sscanf(string(r), "%v-%v", &start, &end)
   if err != nil {
      return 0, 0, err
   }
   return start, end, nil
}