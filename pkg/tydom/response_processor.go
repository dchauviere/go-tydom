package tydom

func (tc *Client) processAsyncResponse() {
	defer tc.wg.Done()

	for {
		select {
		// Process shutdown signal
		case <-tc.shutdown:
			tc.Logger.Debug("processAsyncResponse shutdown")

			return
		case resp := <-tc.asyncResponseQueue:
			tc.Logger.Debug("processing response", "response", resp)

			if hook, exists := tc.asyncResponseRegistry[resp.Header.Get("Transac-Id")]; exists {
				tc.Logger.Debug("executing async hook", "transac-id", resp.Header.Get("Transac-Id"))
				hook(resp)
				delete(tc.asyncResponseRegistry, resp.Header.Get("Transac-Id"))
			}
		}
	}
}
