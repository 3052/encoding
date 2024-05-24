package dash

import (
   "encoding/xml"
   "net/url"
   "strconv"
   "strings"
   "time"
)

type URL struct {
   URL *url.URL
}

type SegmentBase struct {
   IndexRange Range `xml:"indexRange,attr"`
   Initialization struct {
      Range Range `xml:"range,attr"`
   }
}

type Representation struct {
   Bandwidth int64 `xml:"bandwidth,attr"`
   BaseUrl *string
   Codecs *string `xml:"codecs,attr"`
   Height *int64 `xml:"height,attr"`
   ID string `xml:"id,attr"`
   MimeType *string `xml:"mimeType,attr"`
   SegmentBase *SegmentBase
   SegmentTemplate *SegmentTemplate
   Width *int64 `xml:"width,attr"`
   adaptation_set *AdaptationSet
}

func (u *URL) UnmarshalText(text []byte) error {
   return u.URL.UnmarshalBinary(text)
}

func (m *MPD) Unmarshal(data []byte) error {
   err := xml.Unmarshal(data, m)
   if err != nil {
      return err
   }
   for _, period := range m.Period {
      period.mpd = m
      for _, adapt := range period.AdaptationSet {
         adapt.period = period
         for _, represent := range adapt.Representation {
            represent.adaptation_set = adapt
         }
      }
   }
   return nil
}

func (p Period) get_duration() string {
   if v := p.Duration; v != nil {
      return *v
   }
   return p.mpd.MediaPresentationDuration
}

// range values can exceed 32 bits, so make sure to unmarshal to 64 bit
type Range string

func (r Range) Cut() (string, string, bool) {
   return strings.Cut(string(r), "-")
}

func (r Representation) Ext() (string, bool) {
   switch r.get_mime_type() {
   case "audio/mp4":
      return ".m4a", true
   case "video/mp4":
      return ".m4v", true
   }
   return "", false
}

func (r Representation) GetSegmentTemplate() (*SegmentTemplate, bool) {
   if v := r.SegmentTemplate; v != nil {
      return v, true
   }
   if v := r.adaptation_set.SegmentTemplate; v != nil {
      return v, true
   }
   return nil, false
}

func (r Representation) String() string {
   var b []byte
   if v, ok := r.get_width(); ok {
      b = append(b, "width = "...)
      b = strconv.AppendInt(b, v, 10)
   }
   if v, ok := r.get_height(); ok {
      if b != nil {
         b = append(b, '\n')
      }
      b = append(b, "height = "...)
      b = strconv.AppendInt(b, v, 10)
   }
   if b != nil {
      b = append(b, '\n')
   }
   b = append(b, "bandwidth = "...)
   b = strconv.AppendInt(b, r.Bandwidth, 10)
   if v, ok := r.get_codecs(); ok {
      b = append(b, "\ncodecs = "...)
      b = append(b, v...)
   }
   b = append(b, "\ntype = "...)
   b = append(b, r.get_mime_type()...)
   if v := r.adaptation_set.Role; v != nil {
      b = append(b, "\nrole = "...)
      b = append(b, v.Value...)
   }
   if v := r.adaptation_set.Lang; v != nil {
      b = append(b, "\nlang = "...)
      b = append(b, *v...)
   }
   b = append(b, "\nid = "...)
   b = append(b, r.ID...)
   return string(b)
}

func (r Representation) get_codecs() (string, bool) {
   if v := r.Codecs; v != nil {
      return *v, true
   }
   if v := r.adaptation_set.Codecs; v != nil {
      return *v, true
   }
   return "", false
}

func (r Representation) get_height() (int64, bool) {
   if v := r.Height; v != nil {
      return *v, true
   }
   if v := r.adaptation_set.Height; v != nil {
      return *v, true
   }
   return 0, false
}

func (r Representation) get_mime_type() string {
   if v := r.MimeType; v != nil {
      return *v
   }
   return *r.adaptation_set.MimeType
}

func (r Representation) get_width() (int64, bool) {
   if v := r.Width; v != nil {
      return *v, true
   }
   if v := r.adaptation_set.Width; v != nil {
      return *v, true
   }
   return 0, false
}

func (p Period) Seconds() (float64, error) {
   s := strings.TrimPrefix(p.get_duration(), "PT")
   d, err := time.ParseDuration(strings.ToLower(s))
   if err != nil {
      return 0, err
   }
   return d.Seconds(), nil
}

func (r Representation) GetAdaptationSet() *AdaptationSet {
   return r.adaptation_set
}

type AdaptationSet struct {
   Codecs *string `xml:"codecs,attr"`
   Height *int64 `xml:"height,attr"`
   Lang *string `xml:"lang,attr"`
   MimeType *string `xml:"mimeType,attr"`
   Representation []*Representation
   Role *struct {
      Value string `xml:"value,attr"`
   }
   SegmentTemplate *SegmentTemplate
   Width *int64 `xml:"width,attr"`
   period *Period
}

func (a AdaptationSet) GetPeriod() *Period {
   return a.period
}

type Period struct {
   AdaptationSet []*AdaptationSet
   Duration *string `xml:"duration,attr"`
   mpd *MPD
}

func (p Period) GetMpd() *MPD {
   return p.mpd
}

type MPD struct {
   BaseUrl *URL `xml:"BaseURL"`
   MediaPresentationDuration string `xml:"mediaPresentationDuration,attr"`
   Period []*Period
}
