#include <oniguruma.h>

extern int NewOnigRegex( char *pattern, int pattern_length, int option,
                                  OnigRegex *regex, OnigRegion **region, OnigEncoding *encoding, OnigErrorInfo **error_info, char **error_buffer);

extern int SearchOnigRegex( void *str, int str_length, int option,
                                  OnigRegex regex, OnigRegion *region, OnigEncoding encoding, OnigErrorInfo *error_info, char *error_buffer);

extern int IntAt(int *int_array, int index);
