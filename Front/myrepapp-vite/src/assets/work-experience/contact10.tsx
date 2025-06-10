import React, { Fragment } from 'react'
import { ContactProps } from '@/types'
import './contact10.css'

const Contact10: React.FC<ContactProps> = (props) => {
  return (
    <div className="thq-section-padding">
      <div className="contact10-container2 thq-section-max-width">
        <div className="contact10-content1">
          <div className="contact10-content2">
            <h2 className="thq-heading-2">
              {props.heading1 ?? (
                <Fragment>
                  <span>Contact Us</span>
                </Fragment>
              )}
            </h2>
            <p className="thq-body-large">
              {props.content1 ?? (
                <Fragment>
                  <span>Get in touch with us today!</span>
                </Fragment>
              )}
            </p>
          </div>
        </div>
        <div className="contact10-content3">
          <div className="contact10-content4">
            <div className="contact10-content5">
              <span className="thq-body-small">Email</span>
              <a 
                href={`mailto:${props.email1 ? 'kern.zhou1995@gmail.com' : 'contact@example.com'}`}
                className="thq-body-small"
              >
                {props.email1 ?? (
                  <Fragment>
                    <span>contact@example.com</span>
                  </Fragment>
                )}
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Contact10
