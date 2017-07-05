# fake_metron

Fake metron endpoint for testing

## Usage:

```go
  ...
  import "github.com/cloudfoundry/dropsonde/fake_metron"
  ...

  var _ = Describe("Metrics", func() {
    var (
      fakeMetron       *fake_metron.FakeMetron
      fakeMetronClosed chan struct{}
    )
    
    BeforeEach(func() {
      fakeMetron := fake_metron.NewFakeMetron(5000)
      err := fakeMetron.Listen()
      Expect(err).ToNot(HaveOccurred())

      fakeMetronClosed = make(chan struct{})

      go func() {
        defer GinkgoRecover()
        Expect(fakeMetron.Run()).To(Succeed())
        close(fakeMetronClosed)
      }()
    })

    AfterEach(func() {
      Expect(fakeMetron.Stop()).To(Succeed())
      Eventually(fakeMetronClosed).Should(BeClosed())
    })

    It("send metrics to metron", func() {
      //
      // your code that send metrics here
      //

      // verify if the metrics were received:
			var metrics []events.ValueMetric
			Eventually(func() []events.ValueMetric {
				metrics = fakeMetron.ValueMetricsFor("MyMetric")
				return metrics
			}).Should(HaveLen(1))

      Expect(*metrics[0].Name).To(Equal("MyMetric"))
			Expect(*metrics[0].Unit).To(Equal("nanos"))
			Expect(*metrics[0].Value).To(Equal(10.0))
    })
  })

```
