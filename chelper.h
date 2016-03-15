#include <oniguruma.h>

extern int NewOnigRegex(char *pattern, int pattern_length, int option, OnigRegex *regex, OnigRegion **region, char *error_buffer);
extern int SearchOnigRegex(void *str, int str_length, int offset, int option, OnigRegex regex, OnigRegion *region, int *captures, int *numCaptures);

extern int MatchOnigRegex(void *str, int str_length, int offset, int option, OnigRegex regex, OnigRegion *region);

extern int LookupOnigCaptureByName(char *name, int name_length, OnigRegex regex, OnigRegion *region);
extern int GetCaptureNames(OnigRegex regex, void *buffer, int bufferSize, int* groupNumbers);
