package dash

import (
   "math"
   "strconv"
   "strings"
)

// dashif-documents.azurewebsites.net/Guidelines-TimingModel/master/Guidelines-TimingModel.html#addressing-simple-to-explicit
func (s SegmentTemplate) SegmentCount(seconds float64) float64 {
   seconds /= float64(*s.Duration / *s.Timescale)
   return math.Ceil(seconds)
}

// dashif-documents.azurewebsites.net/Guidelines-TimingModel/master/Guidelines-TimingModel.html#timing-sampletimeline
func (s SegmentTemplate) GetTimescale() int {
   if v := s.Timescale; v != nil {
      return *v
   }
   return 1
}

type SegmentTemplate struct {
   Duration *int `xml:"duration,attr"`
   Initialization *string `xml:"initialization,attr"`
   Media string `xml:"media,attr"`
   SegmentTimeline *struct {
      S []struct {
         D int `xml:"d,attr"` // duration
         R *int `xml:"r,attr"` // repeat
      }
   }
   StartNumber *int `xml:"startNumber,attr"`
   Timescale *int `xml:"timescale,attr"`
}

func (s SegmentTemplate) GetInitialization(r Representation) (string, bool) {
   if v := s.Initialization; v != nil {
      return strings.Replace(*v, "$RepresentationID$", r.ID, 1), true
   }
   return "", false
}

func (s SegmentTemplate) GetMedia(r Representation) []string {
   s.Media = strings.Replace(s.Media, "$RepresentationID$", r.ID, 1)
   var (
      media []string
      number int
   )
   if s.StartNumber != nil {
      number = *s.StartNumber
   }
   if s.SegmentTimeline != nil {
      for _, segment := range s.SegmentTimeline.S {
         var repeat int
         if segment.R != nil {
            repeat = *segment.R
         }
         for repeat >= 0 {
            var medium string
            replace := strconv.Itoa(number)
            if s.StartNumber != nil {
               medium = strings.Replace(s.Media, "$Number$", replace, 1)
               number++
            } else {
               medium = strings.Replace(s.Media, "$Time$", replace, 1)
               number += segment.D
            }
            media = append(media, medium)
            repeat--
         }
      }
   } else {
   }
   return media
}
