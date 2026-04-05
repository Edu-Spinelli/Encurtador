package encoder

import hashids "github.com/speps/go-hashids/v2"

type Encoder struct {
	hid *hashids.HashID
}

func New(salt string, pepper string, minLength int) (*Encoder, error) {
	hd := hashids.NewData()
	hd.Salt = salt + pepper
	hd.MinLength = minLength
	hd.Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	hid, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, err
	}

	return &Encoder{hid: hid}, nil
}

func (e *Encoder) Encode(id int64) (string, error) {
	return e.hid.EncodeInt64([]int64{id})
}

func (e *Encoder) Decode(hash string) (int64, error) {
	ids, err := e.hid.DecodeInt64WithError(hash)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}
