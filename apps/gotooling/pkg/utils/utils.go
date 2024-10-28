package utils


func StrInArray(s string, a []string) bool {

  for _, v := range a {
    if s == v {
      return true
    }
  }

  return false
}

func MapKeys(m map[string]*any) []string {
  r := make([]string, 0)

  for k := range m {
    r = append(r, k)
  } 

  return r
}
