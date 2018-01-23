package flag

import (
	"reflect"
	"strings"
	"syscall"

	"encoding/base64"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/misc/strconv"
)

// --------------------------------------------------------------------

/*
Command line:
	METHOD <Command>/<MainParam>/<Switch1>/<Switch1Param>/.../<SwitchN>/<SwitchNParam>
*/
func Parse(ret interface{}, cmdline string) (err error) {

	return ParseValue(reflect.ValueOf(ret), strings.Split(cmdline, "/"), "flag")
}

func ParseEx(ret interface{}, query []string, cate string) (err error) {

	return ParseValue(reflect.ValueOf(ret), query, cate)
}

func ParseValue(v reflect.Value, query []string, cate string) (err error) {

	if v.Kind() != reflect.Ptr {
		err = errors.Info(syscall.EINVAL, "formutil.ParseValue: ret.type != pointer")
		return
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		err = errors.Info(syscall.EINVAL, "formutil.ParseValue: ret.type != struct")
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if sf.Tag == "" { // no tag, skip
			continue
		}
		flagTag := sf.Tag.Get(cate)
		if flagTag == "" {
			continue
		}
		tag, opts, err2 := parseTag(flagTag)
		if err2 != nil {
			err = errors.Info(err2, "Parse struct field:", sf.Name).Detail(err2)
			return
		}
		sfv := v.Field(i)
		fv, ok := findArg(query, tag)
		if opts.fhas {
			if err = setHas(v, sf.Name, ok); err != nil {
				return
			}
		}
		if !ok {
			if !opts.fdefault { // 允许外部设置默认值
				sfv.Set(reflect.Zero(sf.Type))
			}
			continue
		}
		if opts.fbase64 {
			mod := len(fv) & 3
			switch mod {
			case 2:
				fv += "=="
			case 3:
				fv += "="
			}
			b, err2 := base64.URLEncoding.DecodeString(fv)
			if err2 != nil {
				err = errors.Info(err2, "http/flag.ParseValue: parse param", tag).Detail(err2)
				return
			}
			if sfv.Kind() == reflect.Slice {
				if sfv.Type().Elem().Kind() == reflect.Uint8 {
					sfv.SetBytes(b)
				} else {
					err = errors.Info(syscall.EINVAL, "http/flag.ParseValue: parse param", tag)
					return
				}
				continue
			}
			fv = string(b)
		}
		err = strconv.ParseValue(sfv, fv)
		if err != nil {
			err = errors.Info(err, "http/flag.ParseValue: parse param", tag).Detail(err)
			return
		}
	}
	return
}

// --------------------------------------------------------------------

func setHas(v reflect.Value, name string, has bool) (err error) {

	sfHas := v.FieldByName("Has" + name)
	if sfHas.Kind() != reflect.Bool {
		err = errors.New("Struct filed `Has" + name + "` not found or not bool")
		return
	}
	sfHas.SetBool(has)
	return
}

type tagOpts struct {
	fbase64  bool
	fdefault bool
	fhas     bool
}

func parseTag(tag1 string) (tag string, opts tagOpts, err error) {

	if tag1 == "" {
		err = errors.New("Struct field has no tag")
		return
	}

	parts := strings.Split(tag1, ",")
	tag = parts[0]
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "base64":
			opts.fbase64 = true
		case "default":
			opts.fdefault = true
		case "has":
			opts.fhas = true
		default:
			err = errors.New("Unknown tag option: " + parts[i])
			return
		}
	}
	return
}

func findArg(query []string, tag string) (v string, ok bool) {

	if tag == "_" {
		if len(query) < 2 {
			return "", false
		}
		return query[1], true
	}

	for i := 3; i < len(query); i += 2 {
		if query[i-1] == tag {
			return query[i], true
		}
	}
	return "", false
}

// --------------------------------------------------------------------
