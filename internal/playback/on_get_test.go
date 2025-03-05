package playback

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4/seekablebuffer"
	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/test"
	"github.com/stretchr/testify/require"
)

func writeSegment1(t *testing.T, fpath string) {
	init := fmp4.Init{
		Tracks: []*fmp4.InitTrack{
			{
				ID:        1,
				TimeScale: 90000,
				Codec: &fmp4.CodecH264{
					SPS: test.FormatH264.SPS,
					PPS: test.FormatH264.PPS,
				},
			},
			{
				ID:        2,
				TimeScale: 90000,
				Codec: &fmp4.CodecMPEG4Audio{
					Config: mpeg4audio.Config{
						Type:         mpeg4audio.ObjectTypeAACLC,
						SampleRate:   48000,
						ChannelCount: 2,
					},
				},
			},
		},
	}

	var buf1 seekablebuffer.Buffer
	err := init.Marshal(&buf1)
	require.NoError(t, err)

	var buf2 seekablebuffer.Buffer
	parts := fmp4.Parts{
		{
			SequenceNumber: 1,
			Tracks: []*fmp4.PartTrack{{
				ID:       1,
				BaseTime: 0,
				Samples:  []*fmp4.PartSample{},
			}},
		},
		{
			SequenceNumber: 2,
			Tracks: []*fmp4.PartTrack{
				{
					ID:       1,
					BaseTime: 30 * 90000,
					Samples: []*fmp4.PartSample{
						{
							Duration:        30 * 90000,
							IsNonSyncSample: false,
							Payload:         []byte{1, 2},
						},
						{
							Duration:        1 * 90000,
							IsNonSyncSample: false,
							Payload:         []byte{3, 4},
						},
						{
							Duration:        1 * 90000,
							IsNonSyncSample: true,
							Payload:         []byte{5, 6},
						},
					},
				},
				{
					ID:       2,
					BaseTime: 29 * 90000,
					Samples: []*fmp4.PartSample{
						{
							Duration:        30 * 90000,
							IsNonSyncSample: false,
							Payload:         []byte{1, 2},
						},
					},
				},
			},
		},
	}
	err = parts.Marshal(&buf2)
	require.NoError(t, err)

	err = os.WriteFile(fpath, append(buf1.Bytes(), buf2.Bytes()...), 0o644)
	require.NoError(t, err)
}

func writeSegment2(t *testing.T, fpath string) {
	init := fmp4.Init{
		Tracks: []*fmp4.InitTrack{
			{
				ID:        1,
				TimeScale: 90000,
				Codec: &fmp4.CodecH264{
					SPS: test.FormatH264.SPS,
					PPS: test.FormatH264.PPS,
				},
			},
			{
				ID:        2,
				TimeScale: 90000,
				Codec: &fmp4.CodecMPEG4Audio{
					Config: mpeg4audio.Config{
						Type:         mpeg4audio.ObjectTypeAACLC,
						SampleRate:   48000,
						ChannelCount: 2,
					},
				},
			},
		},
	}

	var buf1 seekablebuffer.Buffer
	err := init.Marshal(&buf1)
	require.NoError(t, err)

	var buf2 seekablebuffer.Buffer
	parts := fmp4.Parts{
		{
			SequenceNumber: 3,
			Tracks: []*fmp4.PartTrack{{
				ID:       1,
				BaseTime: 0,
				Samples: []*fmp4.PartSample{
					{
						Duration:        1 * 90000,
						IsNonSyncSample: false,
						Payload:         []byte{7, 8},
					},
					{
						Duration:        1 * 90000,
						IsNonSyncSample: false,
						Payload:         []byte{9, 10},
					},
				},
			}},
		},
		{
			SequenceNumber: 4,
			Tracks: []*fmp4.PartTrack{{
				ID:       1,
				BaseTime: 2 * 90000,
				Samples: []*fmp4.PartSample{
					{
						Duration:        1 * 90000,
						IsNonSyncSample: false,
						Payload:         []byte{11, 12},
					},
				},
			}},
		},
	}
	err = parts.Marshal(&buf2)
	require.NoError(t, err)

	err = os.WriteFile(fpath, append(buf1.Bytes(), buf2.Bytes()...), 0o644)
	require.NoError(t, err)
}

func writeSegment3(t *testing.T, fpath string) {
	init := fmp4.Init{
		Tracks: []*fmp4.InitTrack{
			{
				ID:        1,
				TimeScale: 90000,
				Codec: &fmp4.CodecH264{
					SPS: test.FormatH264.SPS,
					PPS: test.FormatH264.PPS,
				},
			},
		},
	}

	var buf1 seekablebuffer.Buffer
	err := init.Marshal(&buf1)
	require.NoError(t, err)

	var buf2 seekablebuffer.Buffer
	parts := fmp4.Parts{
		{
			SequenceNumber: 1,
			Tracks: []*fmp4.PartTrack{{
				ID:       1,
				BaseTime: 0,
				Samples: []*fmp4.PartSample{
					{
						Duration:        1 * 90000,
						IsNonSyncSample: false,
						Payload:         []byte{13, 14},
					},
				},
			}},
		},
	}
	err = parts.Marshal(&buf2)
	require.NoError(t, err)

	err = os.WriteFile(fpath, append(buf1.Bytes(), buf2.Bytes()...), 0o644)
	require.NoError(t, err)
}

func TestOnGet(t *testing.T) {
	for _, format := range []string{"fmp4", "mp4"} {
		t.Run(format, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "mediamtx-playback")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			err = os.Mkdir(filepath.Join(dir, "mypath"), 0o755)
			require.NoError(t, err)

			writeSegment1(t, filepath.Join(dir, "mypath", "2008-11-07_11-22-00-500000.mp4"))
			writeSegment2(t, filepath.Join(dir, "mypath", "2008-11-07_11-23-02-500000.mp4"))
			writeSegment2(t, filepath.Join(dir, "mypath", "2008-11-07_11-23-04-500000.mp4"))

			s := &Server{
				Address:     "127.0.0.1:9996",
				ReadTimeout: conf.Duration(10 * time.Second),
				PathConfs: map[string]*conf.Path{
					"mypath": {
						Name:       "mypath",
						RecordPath: filepath.Join(dir, "%path/%Y-%m-%d_%H-%M-%S-%f"),
					},
				},
				AuthManager: test.NilAuthManager,
				Parent:      test.NilLogger,
			}
			err = s.Initialize()
			require.NoError(t, err)
			defer s.Close()

			u, err := url.Parse("http://myuser:mypass@localhost:9996/get")
			require.NoError(t, err)

			v := url.Values{}
			v.Set("path", "mypath")
			v.Set("start", time.Date(2008, 11, 0o7, 11, 23, 1, 500000000, time.Local).Format(time.RFC3339Nano))
			v.Set("duration", "3")
			v.Set("format", format)
			u.RawQuery = v.Encode()

			req, err := http.NewRequest(http.MethodGet, u.String(), nil)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, http.StatusOK, res.StatusCode)

			buf, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if format == "fmp4" {
				var parts fmp4.Parts
				err = parts.Unmarshal(buf)
				require.NoError(t, err)

				require.Equal(t, fmp4.Parts{
					{
						SequenceNumber: 0,
						Tracks: []*fmp4.PartTrack{
							{
								ID: 1,
								Samples: []*fmp4.PartSample{
									{
										Duration: 0,
										Payload:  []byte{3, 4},
									},
									{
										Duration:        90000,
										IsNonSyncSample: true,
										Payload:         []byte{5, 6},
									},
								},
							},
						},
					},
					{
						SequenceNumber: 1,
						Tracks: []*fmp4.PartTrack{
							{
								ID:       1,
								BaseTime: 90000,
								Samples: []*fmp4.PartSample{
									{
										Duration: 90000,
										Payload:  []byte{7, 8},
									},
								},
							},
						},
					},
					{
						SequenceNumber: 2,
						Tracks: []*fmp4.PartTrack{
							{
								ID:       1,
								BaseTime: 180000,
								Samples: []*fmp4.PartSample{
									{
										Duration: 90000,
										Payload:  []byte{9, 10},
									},
								},
							},
						},
					},
				}, parts)
			} else {
				require.Equal(t, []byte{
					0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70,
					0x69, 0x73, 0x6f, 0x6d, 0x00, 0x00, 0x00, 0x01,
					0x69, 0x73, 0x6f, 0x6d, 0x69, 0x73, 0x6f, 0x32,
					0x6d, 0x70, 0x34, 0x31, 0x6d, 0x70, 0x34, 0x32,
					0x00, 0x00, 0x04, 0xcf, 0x6d, 0x6f, 0x6f, 0x76,
					0x00, 0x00, 0x00, 0x6c, 0x6d, 0x76, 0x68, 0x64,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
					0xff, 0xff, 0xf8, 0x30, 0x00, 0x01, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x02, 0x5b,
					0x74, 0x72, 0x61, 0x6b, 0x00, 0x00, 0x00, 0x5c,
					0x74, 0x6b, 0x68, 0x64, 0x00, 0x00, 0x00, 0x03,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x0b, 0xb8, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00,
					0x07, 0x80, 0x00, 0x00, 0x04, 0x38, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x24, 0x65, 0x64, 0x74, 0x73,
					0x00, 0x00, 0x00, 0x1c, 0x65, 0x6c, 0x73, 0x74,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x13, 0x88, 0x00, 0x01, 0x5f, 0x90,
					0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0xd3,
					0x6d, 0x64, 0x69, 0x61, 0x00, 0x00, 0x00, 0x20,
					0x6d, 0x64, 0x68, 0x64, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x01, 0x5f, 0x90, 0x00, 0x04, 0x1e, 0xb0,
					0x55, 0xc4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2d,
					0x68, 0x64, 0x6c, 0x72, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x76, 0x69, 0x64, 0x65,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x56, 0x69, 0x64, 0x65,
					0x6f, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x72,
					0x00, 0x00, 0x00, 0x01, 0x7e, 0x6d, 0x69, 0x6e,
					0x66, 0x00, 0x00, 0x00, 0x14, 0x76, 0x6d, 0x68,
					0x64, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x24, 0x64, 0x69, 0x6e, 0x66, 0x00, 0x00, 0x00,
					0x1c, 0x64, 0x72, 0x65, 0x66, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
					0x0c, 0x75, 0x72, 0x6c, 0x20, 0x00, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x01, 0x3e, 0x73, 0x74, 0x62,
					0x6c, 0x00, 0x00, 0x00, 0x96, 0x73, 0x74, 0x73,
					0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x00, 0x86, 0x61, 0x76, 0x63,
					0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x07, 0x80, 0x04, 0x38, 0x00, 0x48, 0x00,
					0x00, 0x00, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x18, 0xff, 0xff, 0x00,
					0x00, 0x00, 0x30, 0x61, 0x76, 0x63, 0x43, 0x01,
					0x42, 0xc0, 0x28, 0x03, 0x01, 0x00, 0x19, 0x67,
					0x42, 0xc0, 0x28, 0xd9, 0x00, 0x78, 0x02, 0x27,
					0xe5, 0x84, 0x00, 0x00, 0x03, 0x00, 0x04, 0x00,
					0x00, 0x03, 0x00, 0xf0, 0x3c, 0x60, 0xc9, 0x20,
					0x01, 0x00, 0x04, 0x08, 0x06, 0x07, 0x08, 0x00,
					0x00, 0x00, 0x18, 0x73, 0x74, 0x74, 0x73, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x04, 0x00, 0x01, 0x5f, 0x90, 0x00,
					0x00, 0x00, 0x1c, 0x73, 0x74, 0x73, 0x73, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00,
					0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x18, 0x63,
					0x74, 0x74, 0x73, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1c, 0x73,
					0x74, 0x73, 0x63, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x24, 0x73, 0x74, 0x73, 0x7a, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x02, 0x00,
					0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00,
					0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x14, 0x73,
					0x74, 0x63, 0x6f, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x04, 0xf9, 0x00,
					0x00, 0x02, 0x00, 0x74, 0x72, 0x61, 0x6b, 0x00,
					0x00, 0x00, 0x5c, 0x74, 0x6b, 0x68, 0x64, 0x00,
					0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00,
					0x00, 0x00, 0x00, 0xff, 0xff, 0xf8, 0x30, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x24, 0x65,
					0x64, 0x74, 0x73, 0x00, 0x00, 0x00, 0x1c, 0x65,
					0x6c, 0x73, 0x74, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0xf2, 0x30, 0x00,
					0x2b, 0xf2, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
					0x00, 0x01, 0x78, 0x6d, 0x64, 0x69, 0x61, 0x00,
					0x00, 0x00, 0x20, 0x6d, 0x64, 0x68, 0x64, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x01, 0x5f, 0x90, 0xff,
					0xfd, 0x40, 0xe0, 0x55, 0xc4, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x2d, 0x68, 0x64, 0x6c, 0x72, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x73,
					0x6f, 0x75, 0x6e, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x53,
					0x6f, 0x75, 0x6e, 0x64, 0x48, 0x61, 0x6e, 0x64,
					0x6c, 0x65, 0x72, 0x00, 0x00, 0x00, 0x01, 0x23,
					0x6d, 0x69, 0x6e, 0x66, 0x00, 0x00, 0x00, 0x10,
					0x73, 0x6d, 0x68, 0x64, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x24,
					0x64, 0x69, 0x6e, 0x66, 0x00, 0x00, 0x00, 0x1c,
					0x64, 0x72, 0x65, 0x66, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x0c,
					0x75, 0x72, 0x6c, 0x20, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0xe7, 0x73, 0x74, 0x62, 0x6c,
					0x00, 0x00, 0x00, 0x67, 0x73, 0x74, 0x73, 0x64,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0x57, 0x6d, 0x70, 0x34, 0x61,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x02, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00,
					0xbb, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x33,
					0x65, 0x73, 0x64, 0x73, 0x00, 0x00, 0x00, 0x00,
					0x03, 0x80, 0x80, 0x80, 0x22, 0x00, 0x02, 0x00,
					0x04, 0x80, 0x80, 0x80, 0x14, 0x40, 0x15, 0x00,
					0x00, 0x00, 0x00, 0x01, 0xf7, 0x39, 0x00, 0x01,
					0xf7, 0x39, 0x05, 0x80, 0x80, 0x80, 0x02, 0x11,
					0x90, 0x06, 0x80, 0x80, 0x80, 0x01, 0x02, 0x00,
					0x00, 0x00, 0x18, 0x73, 0x74, 0x74, 0x73, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x29, 0x32, 0xe0, 0x00,
					0x00, 0x00, 0x18, 0x63, 0x74, 0x74, 0x73, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x1c, 0x73, 0x74, 0x73, 0x63, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x18, 0x73,
					0x74, 0x73, 0x7a, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00,
					0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x14, 0x73,
					0x74, 0x63, 0x6f, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x01, 0x00, 0x00, 0x04, 0xf7, 0x00,
					0x00, 0x00, 0x12, 0x6d, 0x64, 0x61, 0x74, 0x01,
					0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
					0x0a,
				}, buf)
			}
		})
	}
}

func TestOnGetDifferentInit(t *testing.T) {
	dir, err := os.MkdirTemp("", "mediamtx-playback")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = os.Mkdir(filepath.Join(dir, "mypath"), 0o755)
	require.NoError(t, err)

	writeSegment1(t, filepath.Join(dir, "mypath", "2008-11-07_11-22-00-500000.mp4"))
	writeSegment3(t, filepath.Join(dir, "mypath", "2008-11-07_11-23-02-500000.mp4"))

	s := &Server{
		Address:     "127.0.0.1:9996",
		ReadTimeout: conf.Duration(10 * time.Second),
		PathConfs: map[string]*conf.Path{
			"mypath": {
				Name:       "mypath",
				RecordPath: filepath.Join(dir, "%path/%Y-%m-%d_%H-%M-%S-%f"),
			},
		},
		AuthManager: test.NilAuthManager,
		Parent:      test.NilLogger,
	}
	err = s.Initialize()
	require.NoError(t, err)
	defer s.Close()

	u, err := url.Parse("http://myuser:mypass@localhost:9996/get")
	require.NoError(t, err)

	v := url.Values{}
	v.Set("path", "mypath")
	v.Set("start", time.Date(2008, 11, 0o7, 11, 23, 1, 500000000, time.Local).Format(time.RFC3339Nano))
	v.Set("duration", "2")
	v.Set("format", "fmp4")
	u.RawQuery = v.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	buf, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var parts fmp4.Parts
	err = parts.Unmarshal(buf)
	require.NoError(t, err)

	require.Equal(t, fmp4.Parts{
		{
			SequenceNumber: 0,
			Tracks: []*fmp4.PartTrack{
				{
					ID: 1,
					Samples: []*fmp4.PartSample{
						{
							Duration: 0,
							Payload:  []byte{3, 4},
						},
						{
							Duration:        90000,
							IsNonSyncSample: true,
							Payload:         []byte{5, 6},
						},
					},
				},
			},
		},
	}, parts)
}

func TestOnGetNTPCompensation(t *testing.T) {
	dir, err := os.MkdirTemp("", "mediamtx-playback")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = os.Mkdir(filepath.Join(dir, "mypath"), 0o755)
	require.NoError(t, err)

	writeSegment1(t, filepath.Join(dir, "mypath", "2008-11-07_11-22-00-500000.mp4"))
	writeSegment2(t, filepath.Join(dir, "mypath", "2008-11-07_11-23-02-000000.mp4")) // remove 0.5 secs

	s := &Server{
		Address:     "127.0.0.1:9996",
		ReadTimeout: conf.Duration(10 * time.Second),
		PathConfs: map[string]*conf.Path{
			"mypath": {
				Name:       "mypath",
				RecordPath: filepath.Join(dir, "%path/%Y-%m-%d_%H-%M-%S-%f"),
			},
		},
		AuthManager: test.NilAuthManager,
		Parent:      test.NilLogger,
	}
	err = s.Initialize()
	require.NoError(t, err)
	defer s.Close()

	u, err := url.Parse("http://myuser:mypass@localhost:9996/get")
	require.NoError(t, err)

	v := url.Values{}
	v.Set("path", "mypath")
	v.Set("start", time.Date(2008, 11, 0o7, 11, 23, 1, 500000000, time.Local).Format(time.RFC3339Nano))
	v.Set("duration", "3")
	v.Set("format", "fmp4")
	u.RawQuery = v.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	buf, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var parts fmp4.Parts
	err = parts.Unmarshal(buf)
	require.NoError(t, err)

	require.Equal(t, fmp4.Parts{
		{
			SequenceNumber: 0,
			Tracks: []*fmp4.PartTrack{
				{
					ID: 1,
					Samples: []*fmp4.PartSample{
						{
							Duration: 0,
							Payload:  []byte{3, 4},
						},
						{
							Duration:        45000, // 90 - 45
							IsNonSyncSample: true,
							Payload:         []byte{5, 6},
						},
						{
							Duration: 90000,
							Payload:  []byte{7, 8},
						},
					},
				},
			},
		},
		{
			SequenceNumber: 1,
			Tracks: []*fmp4.PartTrack{
				{
					ID:       1,
					BaseTime: 135000,
					Samples: []*fmp4.PartSample{
						{
							Duration: 90000,
							Payload:  []byte{9, 10},
						},
					},
				},
			},
		},
		{
			SequenceNumber: 2,
			Tracks: []*fmp4.PartTrack{
				{
					ID:       1,
					BaseTime: 225000,
					Samples: []*fmp4.PartSample{
						{
							Duration: 90000,
							Payload:  []byte{11, 12},
						},
					},
				},
			},
		},
	}, parts)
}
