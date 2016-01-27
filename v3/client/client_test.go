package client_test

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/cloudfoundry/go-ccapi/v3/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var (
		uaaServer *ghttp.Server
		ccServer  *ghttp.Server
		c         client.Client
	)

	BeforeEach(func() {
		uaaServer = ghttp.NewServer()
		ccServer = ghttp.NewServer()
		c = client.NewClient(ccServer.URL(), uaaServer.URL(), "auth-token")
	})

	AfterEach(func() {
		ccServer.Close()
		uaaServer.Close()
	})

	Describe("GetApplications", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{"Authorization": []string{"auth-token"}}),
					ghttp.VerifyRequest("GET", "/v3/apps"),
					ghttp.RespondWith(http.StatusOK, applicationsJSON),
				),
			)
		})

		It("tries to get applications", func() {
			_, err := c.GetApplications(url.Values{})
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the applications", func() {
			responseJSON, err := c.GetApplications(url.Values{})
			Expect(err).NotTo(HaveOccurred())
			Expect(responseJSON).To(MatchJSON(`
				[{
						"guid": "guid-13552bb6-a866-4e2c-9d47-c2bccc4c35a1",
						"name": "my_app3",
						"desired_state": "STOPPED",
						"total_desired_instances": 0,
						"created_at": "1970-01-01T00:00:03Z",
						"updated_at": null,
						"lifecycle": {
							"type": "buildpack",
							"data": {
								"buildpack": "name-606",
								"stack": "name-607"
							}
						},
						"environment_variables": {
							"magic": "beautiful"
						},
						"links": {}
					},
					{
						"guid": "guid-354f66ff-6618-4f38-aede-a3f8b194c260",
						"name": "my_app2",
						"desired_state": "STOPPED",
						"total_desired_instances": 0,
						"created_at": "1970-01-01T00:00:02Z",
						"updated_at": null,
						"lifecycle": {
							"type": "buildpack",
							"data": {
								"buildpack": "name-604",
								"stack": "name-605"
							}
						},
						"environment_variables": {},
						"links": {}
					}]
				`))
		})

		Context("when the ccServer returns a non-200 response", func() {
			BeforeEach(func() {
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v3/apps"),
						ghttp.RespondWith(http.StatusInternalServerError, ``),
					),
				)
			})

			It("returns an error", func() {
				_, err := c.GetApplications(url.Values{})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when given params", func() {
			BeforeEach(func() {
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v3/apps", "space_guids=space-guid-1,space-guid-2"),
						ghttp.RespondWith(http.StatusOK, applicationsJSON),
					),
				)
			})

			It("uses the params when making the request", func() {
				_, err := c.GetApplications(url.Values{
					"space_guids": []string{"space-guid-1,space-guid-2"},
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the ccServer returns bad JSON", func() {
			BeforeEach(func() {
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v3/apps"),
						ghttp.RespondWith(http.StatusOK, `:bad_json:`),
					),
				)
			})

			It("returns an error", func() {
				_, err := c.GetApplications(url.Values{})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetResource", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{"Authorization": []string{"auth-token"}}),
					ghttp.VerifyRequest("GET", "/the-path"),
					ghttp.RespondWith(http.StatusOK, `{"key": "value"}`),
				),
			)
		})

		It("executes a request for the given path", func() {
			_, err := c.GetResource("/the-path")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the resource from the ccServer", func() {
			responseBody, err := c.GetResource("/the-path")
			Expect(err).NotTo(HaveOccurred())
			Expect(responseBody).To(Equal([]byte(`{"key": "value"}`)))
		})

		Context("when the ccServer returns a non-200 response", func() {
			BeforeEach(func() {
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/the-path"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"error": "some-error"}`),
					),
				)
			})

			It("returns an error", func() {
				_, err := c.GetResource("/the-path")
				Expect(err).To(HaveOccurred())
			})

			It("returns any body the ccServer returns", func() {
				responseBody, _ := c.GetResource("/the-path")
				Expect(responseBody).To(Equal([]byte(`{"error": "some-error"}`)))
			})
		})
	})

	Describe("GetResources", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{"Authorization": []string{"auth-token"}}),
					ghttp.VerifyRequest("GET", "/the-path"),
					ghttp.RespondWith(http.StatusOK, `{
						"pagination": {
							"total_results": 4,
							"first": {
								"href": "/the-path?page=1&per_page=2"
							},
							"last": {
								"href": "/the-path?page=2&per_page=2"
							},
							"next": "/the-path?page=2&per_page=2",
							"previous": null
						},
						"resources": [
							{
								"guid": "resource-1-guid"
							},
							{
								"guid": "resource-2-guid"
							}
						]
						}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/the-path", "page=2&per_page=2"),
					ghttp.RespondWith(http.StatusOK, `{
						"pagination": {
							"total_results": 4,
							"first": {
								"href": "/the-path?page=1&per_page=2"
							},
							"last": {
								"href": "/the-path?page=2&per_page=2"
							},
							"next": null,
							"previous": "/the-path?page=1&per_page=2"
						},
						"resources": [
							{
								"guid": "resource-3-guid"
							},
							{
								"guid": "resource-4-guid"
							}
						]
						}`),
				),
			)
		})

		It("executes a request for each page of resources returned when given a limit of 0", func() {
			_, err := c.GetResources("/the-path", 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
		})

		It("executes only as many requests as necessary to return the requested limit of resources", func() {
			_, err := c.GetResources("/the-path", 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("does not return more resources than are requested", func() {
			responseBytes, err := c.GetResources("/the-path", 3)
			Expect(err).NotTo(HaveOccurred())
			responseStruct := []interface{}{}
			json.Unmarshal(responseBytes, &responseStruct)
			Expect(responseStruct).To(HaveLen(3))
		})

		It("returns an error when given an invalid path", func() {
			_, err := c.GetResources("[%", 0)
			Expect(err).To(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(0))
		})

		Context("when the ccServer responds with bad JSON", func() {
			BeforeEach(func() {
				ccServer.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/the-path"),
						ghttp.RespondWith(http.StatusOK, `:bad_json:`),
					),
				)
			})

			It("returns an error", func() {
				_, err := c.GetResources("/the-path", 0)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("RefreshAuthToken", func() {
		BeforeEach(func() {
			uaaServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.VerifyHeader(map[string][]string{
						"Authorization": []string{"Basic Y2Y6"},
						"accept":        []string{"application/json"},
						"content-type":  []string{"application/x-www-form-urlencoded"},
					}),
				),
			)
		})

		It("tries to refresh the auth token", func() {
			c.RefreshAuthToken()
			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})
	})
})

var performRequestForPathJSON string = `{
  "pagination": {
    "total_results": 1,
    "first": {
      "href": "/v3/apps/guid-1ba69b01-e712-4edc-b7e2-a5a6837c697d/processes?page=1&per_page=50"
    },
    "last": {
      "href": "/v3/apps/guid-1ba69b01-e712-4edc-b7e2-a5a6837c697d/processes?page=1&per_page=50"
    },
    "next": null,
    "previous": null
  },
  "resources": [
    {
      "guid": "4c2d254d-009f-4f47-b789-ece38a72c6ae",
      "type": "web",
      "command": null,
      "instances": 1,
      "memory_in_mb": 1024,
      "disk_in_mb": 1024,
      "created_at": "2015-12-22T18:28:11Z",
      "updated_at": "2015-12-22T18:28:11Z",
      "links": {
        "self": {
          "href": "/v3/processes/4c2d254d-009f-4f47-b789-ece38a72c6ae"
        },
        "scale": {
          "href": "/v3/processes/4c2d254d-009f-4f47-b789-ece38a72c6ae/scale",
          "method": "PUT"
        },
        "app": {
          "href": "/v3/apps/guid-1ba69b01-e712-4edc-b7e2-a5a6837c697d"
        },
        "space": {
          "href": "/v2/spaces/54c6dc91-2992-46bd-b8fa-cb999b370325"
        }
      }
    }
  ]
}`

var applicationsJSON string = `{
"pagination": {
	"total_results": 2,
	"first": {
		"href": "/the-path?page=1&per_page=2"
	},
	"last": {
		"href": "/the-path?page=1&per_page=2"
	},
	"next": null,
	"previous": null
},
"resources": [
  {
    "guid": "guid-13552bb6-a866-4e2c-9d47-c2bccc4c35a1",
    "name": "my_app3",
    "desired_state": "STOPPED",
    "total_desired_instances": 0,
    "created_at": "1970-01-01T00:00:03Z",
    "updated_at": null,
    "lifecycle": {
      "type": "buildpack",
      "data": {
        "buildpack": "name-606",
        "stack": "name-607"
      }
    },
    "environment_variables": {
      "magic": "beautiful"
    },
    "links": {}
  },
  {
    "guid": "guid-354f66ff-6618-4f38-aede-a3f8b194c260",
    "name": "my_app2",
    "desired_state": "STOPPED",
    "total_desired_instances": 0,
    "created_at": "1970-01-01T00:00:02Z",
    "updated_at": null,
    "lifecycle": {
      "type": "buildpack",
      "data": {
        "buildpack": "name-604",
        "stack": "name-605"
      }
    },
    "environment_variables": {},
    "links": {}
  }]}`
