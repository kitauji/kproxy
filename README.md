# kproxy
Simple proxy server implementation written in Go with MITM(man-in-the-middle) feature.

## Usage

You can try kproxy with a sample command.
```bash
cd kproxy/cmd
go build -o proxy
./proxy -port ":9999"
```

If you'd like to try MITM mode for HTTPS requests, at first, create your private CA's certificate(and key). And then install it to the system as the trusted RootCA certificate.

Then run as follows:
```
./proxy -mitm -cacert <CertificatePEMFileFile> -cakey <PrivateKeyPEMFilePath>
```

### Hook Handler
kproxy provides three hook points. For HTTPS requests, they work only when mitm is enabled.

```
	// OnBeforeRequest is a hook point that can be used
	// to customize a request before sending it to the remote host.
	OnBeforeRequest func(r *http.Request) *http.Request

	// OnSendRequest is a hook point that can be used
	// to send a request instead of kproxy.
	// It must return a *http.Response.
	OnSendRequest func(r *http.Request) (*http.Response, error)

	// OnBeforeResponse is a hook point that can be used
	// to customize a response before returning it to the client.
	OnBeforeResponse func(resp *http.Response) *http.Response
```