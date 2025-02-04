package fluentbit

// /*
// #cgo CFLAGS: -I/app/ctrlb-fluent-bit/include -I/app/ctrlb-fluent-bit/build/lib/monkey/include/monkey -I/app/ctrlb-fluent-bit/lib/cmetrics/include -I/app/ctrlb-fluent-bit/lib/ctraces/include -I/app/ctrlb-fluent-bit/lib/mpack-amalgamation-1.1.1/src -I/app/ctrlb-fluent-bit/lib/msgpack-c/include -I/app/ctrlb-fluent-bit/lib/monkey/include/ -I/app/ctrlb-fluent-bit/lib/cfl/include/ -I/app/ctrlb-fluent-bit/lib/cfl/lib/xxhash/ -I/app/ctrlb-fluent-bit/lib/flb_libco/ -I/app/ctrlb-fluent-bit/lib/c-ares-1.33.1/include;
// #cgo LDFLAGS: -L/app/ctrlb-fluent-bit/build/lib  -lfluent-bit -lm -ldl -lpthread
// #include <fluent-bit.h>
// #include <stdlib.h>

// int flb_http_service_set_safe(flb_ctx_t *ctx) {
//     return flb_service_set(ctx, "HTTP_Server", "On", "HTTP_Listen", "0.0.0.0", "HTTP_PORT", "2020", NULL);
// }
// */
// import "C"
// import "unsafe"

// type FlbLibCtx C.struct_flb_lib_ctx
// type CChar C.char
// type FlbCf C.struct_flb_cf
// type FlbConfig C.struct_flb_config

// func (f *FluentBitAdapter) flbReadFromFile(file *C.char) {
// 	((*C.struct_flb_lib_ctx)(f.fluentbitCtx)).config.cf_main = C.flb_read_from_file(((*C.struct_flb_lib_ctx)(f.fluentbitCtx)).config.cf_main, ((*C.struct_flb_lib_ctx)(f.fluentbitCtx)).config, file)
// }

// func (f *FluentBitAdapter) flbCreate() *FlbLibCtx {
// 	return (*FlbLibCtx)(C.flb_create())
// }

// func (f *FluentBitAdapter) flbStart() C.int {
// 	return C.flb_start((*C.struct_flb_lib_ctx)(f.fluentbitCtx))
// }

// func (f *FluentBitAdapter) flbStop() C.int {
// 	return C.flb_stop((*C.struct_flb_lib_ctx)(f.fluentbitCtx))
// }

// func (f *FluentBitAdapter) flbDestroy() {
// 	C.flb_destroy((*C.struct_flb_lib_ctx)(f.fluentbitCtx))
// }

// func (f *FluentBitAdapter) flbDestroyContext(context *FlbLibCtx) {
// 	C.flb_stop((*C.struct_flb_lib_ctx)(context))
// 	C.flb_destroy((*C.struct_flb_lib_ctx)(context))
// }

// func (f *FluentBitAdapter) flbStrdup(s *C.char) *C.char {
// 	return C.flb_strdup(s)
// }

// func (f *FluentBitAdapter) flbCString(s string) *C.char {
// 	return C.CString(s)
// }

// func (f *FluentBitAdapter) flbFreePointer(ptr unsafe.Pointer) {
// 	C.free(ptr)
// }

// func (f *FluentBitAdapter) flbSetHTTPDefaultService() C.int {
// 	return C.flb_http_service_set_safe((*C.struct_flb_lib_ctx)(f.fluentbitCtx))
// }
