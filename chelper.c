#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "chelper.h"

int NewOnigRegex( char *pattern, int pattern_length, int option,
                  OnigRegex *regex, OnigRegion **region, OnigEncoding *encoding, OnigErrorInfo **error_info, char **error_buffer) {
    int ret = ONIG_NORMAL;
    int error_msg_len = 0;

    OnigUChar *pattern_start = (OnigUChar *) pattern;
    OnigUChar *pattern_end = (OnigUChar *) (pattern + pattern_length);

    *error_info = (OnigErrorInfo *) malloc(sizeof(OnigErrorInfo));
    memset(*error_info, 0, sizeof(OnigErrorInfo));

    *encoding = (void*)ONIG_ENCODING_UTF8;

    *error_buffer = (char*) malloc(ONIG_MAX_ERROR_MESSAGE_LEN * sizeof(char));

    memset(*error_buffer, 0, ONIG_MAX_ERROR_MESSAGE_LEN * sizeof(char));

    *region = onig_region_new();

    ret = onig_new(regex, pattern_start, pattern_end, (OnigOptionType)(option), *encoding, OnigDefaultSyntax, *error_info);
  
    if (ret != ONIG_NORMAL) {
        error_msg_len = onig_error_code_to_str((unsigned char*)(*error_buffer), ret, *error_info);
        if (error_msg_len >= ONIG_MAX_ERROR_MESSAGE_LEN) {
            error_msg_len = ONIG_MAX_ERROR_MESSAGE_LEN - 1;
        }
        (*error_buffer)[error_msg_len] = '\0';
    }
    return ret;
}

int SearchOnigRegex( void *str, int str_length, int option,
                  OnigRegex regex, OnigRegion *region, OnigEncoding encoding, OnigErrorInfo *error_info, char *error_buffer) {
    int ret = ONIG_MISMATCH;
    int error_msg_len = 0;

    OnigUChar *str_start = (OnigUChar *) str;
    OnigUChar *str_end = (OnigUChar *) (str_start + str_length);
    OnigUChar *search_start = str_start;
    OnigUChar *search_end = str_end;

    ret = onig_search(regex, str_start, str_end, search_start, search_end, region, option);
    if (ret < 0) {
        error_msg_len = onig_error_code_to_str((unsigned char*)(error_buffer), ret, error_info);
        if (error_msg_len >= ONIG_MAX_ERROR_MESSAGE_LEN) {
            error_msg_len = ONIG_MAX_ERROR_MESSAGE_LEN - 1;
        }
        error_buffer[error_msg_len] = '\0';
    }
    else {
        int i;
        fprintf(stderr, "match at %d\n", ret);
        for (i = 0; i < region->num_regs; i++) {
            fprintf(stderr, "%d: (%d-%d)\n", i, region->beg[i], region->end[i]);
        }
    }
    return ret;
}

int IntAt(int *int_array, int index) {
    return (int)int_array[index];
}

