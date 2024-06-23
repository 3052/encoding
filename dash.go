package dash

import (
   "encoding/xml"
   "fmt"
   "math"
   "strconv"
   "strings"
   "time"
)

func option[T comparable](vals ...T) (T, bool) {
   var zero T
   for _, val := range vals {
      if val != zero {
         return val, true
      }
   }
   return zero, false
}

func (a AdaptationSet) get_role() (*Role, bool) {
   return option(a.Role)
}

type AdaptationSet struct {
   ContentProtection []ContentProtection
   Representation    []*Representation
   period            *Period
   Codecs            string `xml:"codecs,attr"`
   Height            uint64  `xml:"height,attr"`
   Lang              string `xml:"lang,attr"`
   MimeType          string `xml:"mimeType,attr"`
   Width             uint64  `xml:"width,attr"`
   Role              *Role
   SegmentTemplate *SegmentTemplate
}

func (a AdaptationSet) get_lang() (string, bool) {
   return option(a.Lang)
}

func (c ContentProtection) get_pssh() (string, bool) {
   return option(c.Pssh)
}

type ContentProtection struct {
   Pssh        string `xml:"pssh"`
   SchemeIdUri string `xml:"schemeIdUri,attr"`
}

func (d *Duration) UnmarshalText(text []byte) error {
   var err error
   d.D, err = time.ParseDuration(strings.ToLower(
      strings.TrimPrefix(string(text), "PT"),
   ))
   if err != nil {
      return err
   }
   return nil
}

type Duration struct {
   D time.Duration
}

type Mpd struct {
   BaseUrl                   string `xml:"BaseURL"`
   MediaPresentationDuration *Duration `xml:"mediaPresentationDuration,attr"`
   Period                    []*Period
}

func (m *Mpd) Unmarshal(text []byte) error {
   err := xml.Unmarshal(text, m)
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

func (p Period) get_duration() *Duration {
   if v := p.Duration; v != nil {
      return v
   }
   return p.mpd.MediaPresentationDuration
}

type Period struct {
   AdaptationSet []*AdaptationSet
   Duration *Duration `xml:"duration,attr"`
   Id string `xml:"id,attr"`
   mpd           *Mpd
}

// SegmentIndexBox uses:
// unsigned int(32) subsegment_duration;
// but range values can exceed 32 bits
type Range struct {
   Start uint64
   End   uint64
}

func (r Range) MarshalText() ([]byte, error) {
   b := strconv.AppendUint(nil, r.Start, 10)
   b = append(b, '-')
   return strconv.AppendUint(b, r.End, 10), nil
}

func (r *Range) UnmarshalText(text []byte) error {
   // the current testdata always has `-`, so lets assume for now
   start, end, _ := strings.Cut(string(text), "-")
   var err error
   r.Start, err = strconv.ParseUint(start, 10, 64)
   if err != nil {
      return err
   }
   r.End, err = strconv.ParseUint(end, 10, 64)
   if err != nil {
      return err
   }
   return nil
}

func (r Representation) get_segment_template() (*SegmentTemplate, bool) {
   return option(r.SegmentTemplate, r.adaptation_set.SegmentTemplate)
}

func (r Representation) GetMedia() []string {
   template, ok := r.get_segment_template()
   if !ok {
      return nil
   }
   number := template.start()
   template.Media = r.id(template.Media)
   var media []string
   if template.SegmentTimeline != nil {
      for _, segment := range template.SegmentTimeline.S {
         for range 1 + segment.R {
            var medium string
            if strings.Contains(template.Media, "$Time$") {
               medium = template.time(number)
               number += segment.D
            } else {
               medium = template.number(number)
               number++
            }
            media = append(media, medium)
         }
      }
   } else {
      seconds := r.adaptation_set.period.get_duration().D.Seconds()
      for range template.segment_count(seconds) {
         media = append(media, template.number(number))
         number++
      }
   }
   return media
}

func (r Representation) content_protection() []ContentProtection {
   if v := r.ContentProtection; len(v) >= 1 {
      return v
   }
   return r.adaptation_set.ContentProtection
}

func (r Representation) get_mime_type() string {
   if v := r.MimeType; v != "" {
      return v
   }
   return r.adaptation_set.MimeType
}

func (r Representation) String() string {
   var b []byte
   if v, ok := r.get_width(); ok {
      b = append(b, "width = "...)
      b = strconv.AppendUint(b, v, 10)
   }
   if v, ok := r.get_height(); ok {
      if b != nil {
         b = append(b, '\n')
      }
      b = append(b, "height = "...)
      b = strconv.AppendUint(b, v, 10)
   }
   if b != nil {
      b = append(b, '\n')
   }
   b = append(b, "bandwidth = "...)
   b = strconv.AppendUint(b, r.Bandwidth, 10)
   if v, ok := r.get_codecs(); ok {
      b = append(b, "\ncodecs = "...)
      b = append(b, v...)
   }
   b = append(b, "\ntype = "...)
   b = append(b, r.get_mime_type()...)
   if v, ok := r.adaptation_set.get_role(); ok {
      b = append(b, "\nrole = "...)
      b = append(b, v.Value...)
   }
   if v, ok := r.adaptation_set.get_lang(); ok {
      b = append(b, "\nlang = "...)
      b = append(b, v...)
   }
   b = append(b, "\nid = "...)
   b = append(b, r.Id...)
   return string(b)
}

type Representation struct {
   Bandwidth         uint64  `xml:"bandwidth,attr"`
   BaseUrl           string `xml:"BaseURL"`
   ContentProtection []ContentProtection
   Height            uint64  `xml:"height,attr"`
   Id                string `xml:"id,attr"`
   MimeType          string `xml:"mimeType,attr"`
   Width             uint64  `xml:"width,attr"`
   adaptation_set    *AdaptationSet
   Codecs            string `xml:"codecs,attr"`
   SegmentBase       *SegmentBase
   SegmentTemplate   *SegmentTemplate
}

func (r Representation) id(value string) string {
   return strings.Replace(value, "$RepresentationID$", r.Id, 1)
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

func (r Representation) Widevine() (string, bool) {
   for _, p := range r.content_protection() {
      if p.SchemeIdUri == "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed" {
         return p.get_pssh()
      }
   }
   return "", false
}

func (r Representation) Initialization() (string, bool) {
   if v, ok := r.get_segment_template(); ok {
      if v, ok := v.get_initialization(); ok {
         return r.id(v), true
      }
   }
   return "", false
}

func (r Representation) get_width() (uint64, bool) {
   return option(r.Width, r.adaptation_set.Width)
}

func (r Representation) get_height() (uint64, bool) {
   return option(r.Height, r.adaptation_set.Height)
}

func (r Representation) get_codecs() (string, bool) {
   return option(r.Codecs, r.adaptation_set.Codecs)
}

type Role struct {
   Value string `xml:"value,attr"`
}

type SegmentBase struct {
   Initialization struct {
      Range Range `xml:"range,attr"`
   }
   IndexRange Range `xml:"indexRange,attr"`
}

func (s SegmentTemplate) number(value uint) string {
   f := strings.Replace(s.Media, "$Number$", "%d", 1)
   f = strings.Replace(f, "$Number%02d$", "%02d", 1)
   f = strings.Replace(f, "$Number%03d$", "%03d", 1)
   f = strings.Replace(f, "$Number%04d$", "%04d", 1)
   f = strings.Replace(f, "$Number%05d$", "%05d", 1)
   f = strings.Replace(f, "$Number%06d$", "%06d", 1)
   f = strings.Replace(f, "$Number%07d$", "%07d", 1)
   f = strings.Replace(f, "$Number%08d$", "%08d", 1)
   f = strings.Replace(f, "$Number%09d$", "%09d", 1)
   return fmt.Sprintf(f, value)
}

func (s SegmentTemplate) time(value uint) string {
   f := strings.Replace(s.Media, "$Time$", "%d", 1)
   return fmt.Sprintf(f, value)
}

func (s SegmentTemplate) get_initialization() (string, bool) {
   return option(s.Initialization)
}

func (s SegmentTemplate) start() uint {
   if v := s.PresentationTimeOffset; v >= 1 {
      return v
   }
   return s.StartNumber
}

// dashif-documents.azurewebsites.net/Guidelines-TimingModel/master/Guidelines-TimingModel.html#timing-sampletimeline
func (s SegmentTemplate) get_timescale() uint64 {
   if v := s.Timescale; v >= 1 {
      return v
   }
   return 1
}

// dashif-documents.azurewebsites.net/Guidelines-TimingModel/master/Guidelines-TimingModel.html#addressing-simple-to-explicit
func (s SegmentTemplate) segment_count(seconds float64) uint64 {
   seconds /= float64(s.Duration) / float64(s.get_timescale())
   return uint64(math.Ceil(seconds))
}

type SegmentTemplate struct {
   Duration uint64 `xml:"duration,attr"`
   Initialization string `xml:"initialization,attr"`
   Media string `xml:"media,attr"`
   StartNumber uint `xml:"startNumber,attr"`
   PresentationTimeOffset uint `xml:"presentationTimeOffset,attr"`
   Timescale uint64 `xml:"timescale,attr"`
   SegmentTimeline *struct {
      S []struct {
         D uint `xml:"d,attr"` // duration
         R uint `xml:"r,attr"` // repeat
      }
   }
}
