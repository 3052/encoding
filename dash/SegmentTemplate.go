package dash

import (
   "errors"
   "math"
   "net/url"
   "strconv"
   "strings"
)

func (r Representation) GetSegmentTemplate() (*SegmentTemplate, bool) {
   if v := r.SegmentTemplate; v != nil {
      return v, true
   }
   if v := r.adaptation_set.SegmentTemplate; v != nil {
      return v, true
   }
   return nil, false
}

func (s SegmentTemplate) GetMedia(r *Representation) ([]string, error) {
   s.Media = strings.Replace(s.Media, "$RepresentationID$", r.ID, 1)
   var number int
   if s.StartNumber != nil {
      number = *s.StartNumber
   }
   var media []string
   if s.SegmentTimeline != nil {
      for _, segment := range s.SegmentTimeline.S {
         var repeat int
         if segment.R != nil {
            repeat = *segment.R
         }
         for range 1 + repeat {
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
         }
      }
   } else {
      seconds, err := r.adaptation_set.period.Seconds()
      if err != nil {
         return nil, err
      }
      for range int(s.segment_count(seconds)) {
         replace := strconv.Itoa(number)
         medium := strings.Replace(s.Media, "$Number$", replace, 1)
         media = append(media, medium)
         number++
      }
   }
   return media, nil
}

func (s SegmentTemplate) GetInitialization(r *Representation) (*url.URL, error) {
   if v := s.Initialization; v != nil {
      u, err := url.Parse(r.adaptation_set.period.mpd.BaseURL)
      if err != nil {
         return nil, err
      }
      return u.Parse(strings.Replace(*v, "$RepresentationID$", r.ID, 1))
   }
   return nil, errors.New("GetInitialization")
}

type SegmentTemplate struct {
   Duration *float64 `xml:"duration,attr"`
   Initialization *string `xml:"initialization,attr"`
   Media string `xml:"media,attr"`
   SegmentTimeline *struct {
      S []struct {
         D int `xml:"d,attr"` // duration
         R *int `xml:"r,attr"` // repeat
      }
   }
   StartNumber *int `xml:"startNumber,attr"`
   Timescale *float64 `xml:"timescale,attr"`
}

// dashif-documents.azurewebsites.net/Guidelines-TimingModel/master/Guidelines-TimingModel.html#timing-sampletimeline
func (s SegmentTemplate) get_timescale() float64 {
   if v := s.Timescale; v != nil {
      return *v
   }
   return 1
}

// dashif-documents.azurewebsites.net/Guidelines-TimingModel/master/Guidelines-TimingModel.html#addressing-simple-to-explicit
func (s SegmentTemplate) segment_count(seconds float64) float64 {
   seconds /= *s.Duration / s.get_timescale()
   return math.Ceil(seconds)
}
