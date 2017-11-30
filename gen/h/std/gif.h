#ifndef WUFFS_GIF_H
#define WUFFS_GIF_H

// Code generated by wuffs-c. DO NOT EDIT.

#ifndef WUFFS_BASE_HEADER_H
#define WUFFS_BASE_HEADER_H

// Copyright 2017 The Wuffs Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <stdbool.h>
#include <stdint.h>
#include <string.h>

// Wuffs requires a word size of at least 32 bits because it assumes that
// converting a u32 to usize will never overflow. For example, the size of a
// decoded image is often represented, explicitly or implicitly in an image
// file, as a u32, and it is convenient to compare that to a buffer size.
//
// Similarly, the word size is at most 64 bits because it assumes that
// converting a usize to u64 will never overflow.
#if __WORDSIZE < 32
#error "Wuffs requires a word size of at least 32 bits"
#elif __WORDSIZE > 64
#error "Wuffs requires a word size of at most 64 bits"
#endif

// WUFFS_VERSION is the major.minor version number as a uint32. The major
// number is the high 16 bits. The minor number is the low 16 bits.
//
// The intention is to bump the version number at least on every API / ABI
// backwards incompatible change.
//
// For now, the API and ABI are simply unstable and can change at any time.
//
// TODO: don't hard code this in base-header.h.
#define WUFFS_VERSION (0x00001)

// wuffs_base__slice_u8 is a 1-dimensional buffer (a pointer and length).
//
// A value with all fields NULL or zero is a valid, empty slice.
typedef struct {
  uint8_t* ptr;
  size_t len;
} wuffs_base__slice_u8;

// wuffs_base__buf1 is a 1-dimensional buffer (a pointer and length), plus
// additional indexes into that buffer, plus an opened / closed flag.
//
// A value with all fields NULL or zero is a valid, empty buffer.
typedef struct {
  uint8_t* ptr;  // Pointer.
  size_t len;    // Length.
  size_t wi;     // Write index. Invariant: wi <= len.
  size_t ri;     // Read  index. Invariant: ri <= wi.
  bool closed;   // No further writes are expected.
} wuffs_base__buf1;

// wuffs_base__limit1 provides a limited view of a 1-dimensional byte stream:
// its first N bytes. That N can be greater than a buffer's current read or
// write capacity. N decreases naturally over time as bytes are read from or
// written to the stream.
//
// A value with all fields NULL or zero is a valid, unlimited view.
typedef struct wuffs_base__limit1 {
  uint64_t* ptr_to_len;             // Pointer to N.
  struct wuffs_base__limit1* next;  // Linked list of limits.
} wuffs_base__limit1;

typedef struct {
  // TODO: move buf into private_impl? As it is, it looks like users can modify
  // the buf field to point to a different buffer, which can turn the limit and
  // mark fields into dangling pointers.
  wuffs_base__buf1* buf;
  // Do not access the private_impl's fields directly. There is no API/ABI
  // compatibility or safety guarantee if you do so.
  struct {
    wuffs_base__limit1 limit;
    uint8_t* mark;
  } private_impl;
} wuffs_base__reader1;

typedef struct {
  // TODO: move buf into private_impl? As it is, it looks like users can modify
  // the buf field to point to a different buffer, which can turn the limit and
  // mark fields into dangling pointers.
  wuffs_base__buf1* buf;
  // Do not access the private_impl's fields directly. There is no API/ABI
  // compatibility or safety guarantee if you do so.
  struct {
    wuffs_base__limit1 limit;
    uint8_t* mark;
  } private_impl;
} wuffs_base__writer1;

#endif  // WUFFS_BASE_HEADER_H

#ifdef __cplusplus
extern "C" {
#endif

// ---------------- Status Codes

// Status codes are int32_t values:
//  - the sign bit indicates a non-recoverable status code: an error
//  - bits 10-30 hold the packageid: a namespace
//  - bits 8-9 are reserved
//  - bits 0-7 are a package-namespaced numeric code
//
// Do not manipulate these bits directly. Use the API functions such as
// wuffs_gif__status__is_error instead.
typedef int32_t wuffs_gif__status;

#define wuffs_gif__packageid 1017222  // 0x000f8586

#define WUFFS_GIF__STATUS_OK 0                               // 0x00000000
#define WUFFS_GIF__ERROR_BAD_WUFFS_VERSION -2147483647       // 0x80000001
#define WUFFS_GIF__ERROR_BAD_RECEIVER -2147483646            // 0x80000002
#define WUFFS_GIF__ERROR_BAD_ARGUMENT -2147483645            // 0x80000003
#define WUFFS_GIF__ERROR_INITIALIZER_NOT_CALLED -2147483644  // 0x80000004
#define WUFFS_GIF__ERROR_INVALID_I_O_OPERATION -2147483643   // 0x80000005
#define WUFFS_GIF__ERROR_CLOSED_FOR_WRITES -2147483642       // 0x80000006
#define WUFFS_GIF__ERROR_UNEXPECTED_EOF -2147483641          // 0x80000007
#define WUFFS_GIF__SUSPENSION_SHORT_READ 8                   // 0x00000008
#define WUFFS_GIF__SUSPENSION_SHORT_WRITE 9                  // 0x00000009

#define WUFFS_GIF__ERROR_BAD_GIF_BLOCK -1105848320            // 0xbe161800
#define WUFFS_GIF__ERROR_BAD_GIF_EXTENSION_LABEL -1105848319  // 0xbe161801
#define WUFFS_GIF__ERROR_BAD_GIF_HEADER -1105848318           // 0xbe161802
#define WUFFS_GIF__ERROR_BAD_LZW_LITERAL_WIDTH -1105848317    // 0xbe161803
#define WUFFS_GIF__ERROR_INTERNAL_ERROR_INCONSISTENT_LIMITED_READ \
  -1105848316  // 0xbe161804
#define WUFFS_GIF__ERROR_TODO_UNSUPPORTED_LOCAL_COLOR_TABLE \
  -1105848315                                                      // 0xbe161805
#define WUFFS_GIF__ERROR_LZW_CODE_IS_OUT_OF_RANGE -1105848314      // 0xbe161806
#define WUFFS_GIF__ERROR_LZW_PREFIX_CHAIN_IS_CYCLICAL -1105848313  // 0xbe161807

bool wuffs_gif__status__is_error(wuffs_gif__status s);

const char* wuffs_gif__status__string(wuffs_gif__status s);

// ---------------- Public Consts

// ---------------- Structs

typedef struct {
  // Do not access the private_impl's fields directly. There is no API/ABI
  // compatibility or safety guarantee if you do so. Instead, use the
  // wuffs_gif__lzw_decoder__etc functions.
  //
  // In C++, these fields would be "private", but C does not support that.
  //
  // It is a struct, not a struct*, so that it can be stack allocated.
  struct {
    wuffs_gif__status status;
    uint32_t magic;

    uint32_t f_literal_width;
    uint8_t f_stack[4096];
    uint8_t f_suffixes[4096];
    uint16_t f_prefixes[4096];

    struct {
      uint32_t coro_susp_point;
      uint32_t v_clear_code;
      uint32_t v_end_code;
      uint32_t v_save_code;
      uint32_t v_prev_code;
      uint32_t v_width;
      uint32_t v_bits;
      uint32_t v_n_bits;
      uint32_t v_code;
      uint32_t v_s;
      uint32_t v_c;
      uint64_t v_n_copied;
    } c_decode[1];
  } private_impl;
} wuffs_gif__lzw_decoder;

typedef struct {
  // Do not access the private_impl's fields directly. There is no API/ABI
  // compatibility or safety guarantee if you do so. Instead, use the
  // wuffs_gif__decoder__etc functions.
  //
  // In C++, these fields would be "private", but C does not support that.
  //
  // It is a struct, not a struct*, so that it can be stack allocated.
  struct {
    wuffs_gif__status status;
    uint32_t magic;

    uint32_t f_width;
    uint32_t f_height;
    uint8_t f_background_color_index;
    uint8_t f_gct[768];
    wuffs_gif__lzw_decoder f_lzw;

    struct {
      uint32_t coro_susp_point;
      uint8_t v_c;
    } c_decode[1];
    struct {
      uint32_t coro_susp_point;
      uint8_t v_c[6];
      uint32_t v_i;
    } c_decode_header[1];
    struct {
      uint32_t coro_susp_point;
      uint8_t v_c[7];
      uint32_t v_i;
      uint32_t v_gct_size;
    } c_decode_lsd[1];
    struct {
      uint32_t coro_susp_point;
      uint8_t v_label;
      uint8_t v_block_size;
      uint64_t scratch;
    } c_decode_extension[1];
    struct {
      uint32_t coro_susp_point;
      uint8_t v_c[9];
      uint32_t v_i;
      bool v_interlace;
      uint8_t v_lw;
      uint64_t v_block_size;
      wuffs_gif__status v_z;
    } c_decode_id[1];
  } private_impl;
} wuffs_gif__decoder;

// ---------------- Public Initializer Prototypes

// wuffs_gif__lzw_decoder__initialize is an initializer function.
//
// It should be called before any other wuffs_gif__lzw_decoder__* function.
//
// Pass WUFFS_VERSION and 0 for wuffs_version and for_internal_use_only.
void wuffs_gif__lzw_decoder__initialize(wuffs_gif__lzw_decoder* self,
                                        uint32_t wuffs_version,
                                        uint32_t for_internal_use_only);

// wuffs_gif__decoder__initialize is an initializer function.
//
// It should be called before any other wuffs_gif__decoder__* function.
//
// Pass WUFFS_VERSION and 0 for wuffs_version and for_internal_use_only.
void wuffs_gif__decoder__initialize(wuffs_gif__decoder* self,
                                    uint32_t wuffs_version,
                                    uint32_t for_internal_use_only);

// ---------------- Public Function Prototypes

wuffs_gif__status wuffs_gif__decoder__decode(wuffs_gif__decoder* self,
                                             wuffs_base__writer1 a_dst,
                                             wuffs_base__reader1 a_src);

void wuffs_gif__lzw_decoder__set_literal_width(wuffs_gif__lzw_decoder* self,
                                               uint32_t a_lw);

wuffs_gif__status wuffs_gif__lzw_decoder__decode(wuffs_gif__lzw_decoder* self,
                                                 wuffs_base__writer1 a_dst,
                                                 wuffs_base__reader1 a_src);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // WUFFS_GIF_H
