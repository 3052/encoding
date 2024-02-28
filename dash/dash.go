package dash

import (
   "encoding/base64"
   "encoding/hex"
   "errors"
   "fmt"
   "strings"
)

// media presentation description
// wikipedia.org/wiki/Dynamic_Adaptive_Streaming_over_HTTP
type MPD struct {
   Period []Period
}

func (m MPD) Contains(f func(Pointer) bool) bool {
   for _, period := range m.Period {
      for _, adapt := range period.AdaptationSet {
         for _, represent := range adapt.Representation {
            var p Pointer
            p.AdaptationSet = &adapt
            p.Period = &period
            p.Representation = &represent
            if f(p) {
               return true
            }
         }
      }
   }
   return false
}

func (m MPD) Visit(f func(Pointer)) {
   m.Contains(func(p Pointer) bool {
      f(p)
      return false
   })
}

func (m MPD) String() string {
   var b []byte
   m.Visit(func(p Pointer) {
      if b != nil {
         b = append(b, "\n\n"...)
      }
      var c []byte
      if v := p.Representation.Width; v >= 1 {
         c = fmt.Append(c, "width = ", v)
      }
      if v := p.Representation.Height; v >= 1 {
         if c != nil {
            c = append(c, '\n')
         }
         c = fmt.Append(c, "height = ", v)
      }
      if c != nil {
         c = append(c, '\n')
      }
      c = fmt.Append(c, "bandwidth = ", p.Representation.Bandwidth)
      if v, ok := p.codecs(); ok {
         c = fmt.Append(c, "\ncodecs = ", v)
      }
      c = fmt.Append(c, "\ntype = ", p.mime_type())
      if v := p.AdaptationSet.Role; v != nil {
         c = fmt.Append(c, "\nrole = ", v.Value)
      }
      if v := p.AdaptationSet.Lang; v != "" {
         c = fmt.Append(c, "\nlang = ", v)
      }
      c = fmt.Append(c, "\nid = ", p.Representation.ID)
      if v := p.Period.ID; v != "" {
         c = fmt.Append(c, "\nperiod = ", v)
      }
      b = append(b, c...)
   })
   return string(b)
}
type AdaptationSet struct {
   // this might not exist, or might be under Representation
   Codecs string `xml:"codecs,attr"`
   // this might be under Representation
   ContentProtection []ContentProtection
   // this might not exist
   Lang string `xml:"lang,attr"`
   // this might be under Representation
   MimeType string `xml:"mimeType,attr"`
   Representation []Representation
   // this might not exist
   Role *struct {
      Value string `xml:"value,attr"`
   }
   // this might not exist, or might be under Representation
   SegmentTemplate *SegmentTemplate
}

type ContentProtection struct {
   SchemeIdUri string `xml:"schemeIdUri,attr"`
   // this might not exist
   Default_KID string `xml:"default_KID,attr"`
   // this might not exist
   PSSH string `xml:"pssh"`
}

type Period struct {
   AdaptationSet []AdaptationSet
   ID string `xml:"id,attr"`
}

type Pointer struct {
   AdaptationSet *AdaptationSet
   Period *Period
   Representation *Representation
}

func (p Pointer) Default_KID() ([]byte, error) {
   for _, c := range p.content_protection() {
      if c.SchemeIdUri == "urn:mpeg:dash:mp4protection:2011" {
         c.Default_KID = strings.ReplaceAll(c.Default_KID, "-", "")
         return hex.DecodeString(c.Default_KID)
      }
   }
   return nil, errors.New("Pointer.Default_KID")
}

func (p Pointer) Ext() (string, bool) {
   switch p.mime_type() {
   case "audio/mp4":
      return ".m4a", true
   case "video/mp4":
      return ".m4v", true
   }
   return "", false
}

func (p Pointer) Initialization() (string, bool) {
   if st, ok := p.segment_template(); ok {
      if i := st.Initialization; i != "" {
         i = strings.Replace(i, "$RepresentationID$", p.Representation.ID, 1)
         return i, true
      }
   }
   return "", false
}

func (p Pointer) Media() []string {
   st, ok := p.segment_template()
   if !ok {
      return nil
   }
   replace := func(s, old string) string {
      s = strings.Replace(s, "$RepresentationID$", p.Representation.ID, 1)
      return strings.Replace(s, old, fmt.Sprint(st.StartNumber), 1)
   }
   var media []string
   for _, segment := range st.SegmentTimeline.S {
      for segment.R >= 0 {
         var medium string
         if strings.Contains(st.Media, "$Time$") {
            medium = replace(st.Media, "$Time$")
            st.StartNumber += segment.D
         } else {
            medium = replace(st.Media, "$Number$")
            st.StartNumber++
         }
         media = append(media, medium)
         segment.R--
      }
   }
   return media
}

func (p Pointer) PSSH() ([]byte, error) {
   for _, c := range p.content_protection() {
      if c.SchemeIdUri == "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed" {
         if c.PSSH != "" {
            return base64.StdEncoding.DecodeString(c.PSSH)
         }
      }
   }
   return nil, errors.New("Pointer.PSSH")
}

func (p Pointer) codecs() (string, bool) {
   if v := p.AdaptationSet.Codecs; v != "" {
      return v, true
   }
   if v := p.Representation.Codecs; v != "" {
      return v, true
   }
   return "", false
}

func (p Pointer) content_protection() []ContentProtection {
   if v := p.AdaptationSet.ContentProtection; v != nil {
      return v
   }
   return p.Representation.ContentProtection
}

func (p Pointer) mime_type() string {
   if v := p.AdaptationSet.MimeType; v != "" {
      return v
   }
   return p.Representation.MimeType
}

func (p Pointer) segment_template() (*SegmentTemplate, bool) {
   if v := p.AdaptationSet.SegmentTemplate; v != nil {
      return v, true
   }
   if v := p.Representation.SegmentTemplate; v != nil {
      return v, true
   }
   return nil, false
}

type Range string

func (r Range) Scan() (int, int, error) {
   var start, end int
   _, err := fmt.Sscanf(string(r), "%v-%v", &start, &end)
   if err != nil {
      return 0, 0, err
   }
   return start, end, nil
}

type Representation struct {
   Bandwidth int `xml:"bandwidth,attr"`
   ID string `xml:"id,attr"`
   // this might not exist
   BaseURL string
   // this might not exist, or might be under AdaptationSet
   Codecs string `xml:"codecs,attr"`
   // this might be under AdaptationSet
   ContentProtection []ContentProtection
   // this might not exist
   Height int `xml:"height,attr"`
   // this might be under AdaptationSet
   MimeType string `xml:"mimeType,attr"`
   // this might not exist
   SegmentBase *struct {
      Initialization struct {
         Range Range `xml:"range,attr"`
      }
      IndexRange Range `xml:"indexRange,attr"`
   }
   // this might not exist, or might be under AdaptationSet
   SegmentTemplate *SegmentTemplate
   // this might not exist
   Width int `xml:"width,attr"`
}

type SegmentTemplate struct {
   Media string `xml:"media,attr"`
   SegmentTimeline struct {
      S []struct {
         // duration
         D int `xml:"d,attr"`
         // repeat. this may not exist
         R int `xml:"r,attr"`
      }
   }
   StartNumber int `xml:"startNumber,attr"`
   // this may not exist
   Initialization string `xml:"initialization,attr"`
}
