package dash

import (
   "encoding/base64"
   "errors"
   "strconv"
   "strings"
)

func (p Pointer) PSSH() ([]byte, error) {
   for _, c := range p.contentProtection() {
      if c.SchemeIdUri == "urn:uuid:edef8ba9-79d6-4ace-a3c8-27dcd51d21ed" {
         return base64.StdEncoding.DecodeString(c.PSSH)
      }
   }
   return nil, errors.New("Pointer.PSSH")
}

func (m MPD) Every(f func(Pointer)) {
   m.Some(func(p Pointer) bool {
      f(p)
      return true
   })
}

func (m MPD) Some(f func(Pointer) bool) {
   for _, period := range m.Period {
      for _, adapt := range period.AdaptationSet {
         for _, represent := range adapt.Representation {
            var p Pointer
            p.AdaptationSet = adapt
            p.Period = period
            p.Representation = represent
            if !f(p) {
               return
            }
         }
      }
   }
}

type Pointer struct {
   AdaptationSet *AdaptationSet
   Period *Period
   Representation *Representation
}

func (p Pointer) Codecs() string {
   if a := p.AdaptationSet; a.Codecs != "" {
      return a.Codecs
   }
   return p.Representation.Codecs
}

func (p Pointer) Ext() (string, bool) {
   switch p.MimeType() {
   case "audio/mp4":
      return ".m4a", true
   case "video/mp4":
      return ".m4v", true
   }
   return "", false
}

func (p Pointer) Initialization() (string, bool) {
   if st := p.segmentTemplate(); st != nil {
      if i := st.Initialization; i != "" {
         i = strings.Replace(i, "$RepresentationID$", p.Representation.ID, 1)
         return i, true
      }
   }
   return "", false
}

// return a slice so we can measure progress
func (p Pointer) Media() []string {
   replace := func(s string, i int) string {
      s = strings.Replace(s, "$RepresentationID$", p.Representation.ID, 1)
      return strings.Replace(s, "$Number$", strconv.Itoa(i), 1)
   }
   var media []string
   if st := p.segmentTemplate(); st != nil {
      for _, segment := range st.SegmentTimeline.S {
         for segment.R >= 0 {
            medium := replace(st.Media, st.StartNumber)
            media = append(media, medium)
            segment.R--
            st.StartNumber++
         }
      }
   }
   return media
}

func (p Pointer) MimeType() string {
   if a := p.AdaptationSet; a.MimeType != "" {
      return a.MimeType
   }
   return p.Representation.MimeType
}

func (p Pointer) contentProtection() []ContentProtection {
   if a := p.AdaptationSet; a.ContentProtection != nil {
      return a.ContentProtection
   }
   return p.Representation.ContentProtection
}

func (p Pointer) segmentTemplate() *SegmentTemplate {
   if a := p.AdaptationSet; a.SegmentTemplate != nil {
      return a.SegmentTemplate
   }
   return p.Representation.SegmentTemplate
}
