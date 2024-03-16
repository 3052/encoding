package dash

import (
   "encoding/base64"
   "encoding/hex"
   "fmt"
   "strings"
)

type ContentProtection struct {
   SchemeIdUri string `xml:"schemeIdUri,attr"`
   // this might not exist
   Default_KID Default_KID `xml:"default_KID,attr"`
   // this might not exist
   PSSH PSSH `xml:"pssh"`
}

type Default_KID string

func (d Default_KID) Decode() ([]byte, error) {
   s := strings.ReplaceAll(string(d), "-", "")
   return hex.DecodeString(s)
}

type PSSH string

func (p PSSH) Decode() ([]byte, error) {
   return base64.StdEncoding.DecodeString(string(p))
}

func (r Representation) Default_KID() (Default_KID, bool) {
   for _, c := range r.content_protection() {
      if c.SchemeIdUri == "urn:mpeg:dash:mp4protection:2011" {
         return c.Default_KID, true
      }
   }
   return "", false
}

func (r Representation) PSSH() (PSSH, bool) {
   for _, c := range r.content_protection() {
      if c.SchemeIdUri == "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed" {
         if c.PSSH != "" {
            return c.PSSH, true
         }
      }
   }
   return "", false
}

type Representation struct {
   adaptation_set *adaptation_set
   Bandwidth int64 `xml:"bandwidth,attr"`
   ID string `xml:"id,attr"`
   // this might not exist
   BaseURL string
   // this might not exist, or might be under AdaptationSet
   Codecs string `xml:"codecs,attr"`
   // this might be under AdaptationSet
   ContentProtection []ContentProtection
   // this might not exist
   Height int64 `xml:"height,attr"`
   // this might be under AdaptationSet
   MimeType string `xml:"mimeType,attr"`
   // this might not exist
   SegmentBase *struct {
      Initialization struct {
         Range RawRange `xml:"range,attr"`
      }
      IndexRange RawRange `xml:"indexRange,attr"`
   }
   // this might not exist, or might be under AdaptationSet
   SegmentTemplate *SegmentTemplate
   // this might not exist
   Width int64 `xml:"width,attr"`
}

type RawRange string

type Range struct {
   Start uint64
   End uint64
}

func (r RawRange) Scan() (*Range, error) {
   var v Range
   _, err := fmt.Sscanf(string(r), "%v-%v", &v.Start, &v.End)
   if err != nil {
      return nil, err
   }
   return &v, nil
}