import React, { Fragment, useEffect, useState } from 'react';
import { Helmet } from 'react-helmet';
import { getWorkExperience, WorkExperienceDataResponse } from '@/api/content';
import './ZsWorkExperience.css';

import Hero17 from '@/assets/work-experience/hero17';
import Features24 from '@/assets/work-experience/features24';
import Features25 from '@/assets/work-experience/features25';
import Steps2 from '@/assets/work-experience/steps2';
import Contact10 from '@/assets/work-experience/contact10';

import '@/assets/work-experience/global-style.css';
import '@/assets/work-experience/home.css';

/* helper: 用 React.Fragment + span 包裹字符串，匹配组件预期的 ReactElement 类型 */
const wrap = (text?: string) => (
  <Fragment><span>{text ?? ''}</span></Fragment>
);

const ZsWorkExperience: React.FC = () => {
  const [data, setData] = useState<WorkExperienceDataResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getWorkExperience()
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load'));
  }, []);

  if (error) {
    return <div className="home-container"><p style={{color:'var(--color-gold)',padding:'6rem 2rem'}}>{error}</p></div>;
  }

  if (!data) {
    return <div className="home-container" />;
  }

  return (
    <div className="home-container">
      <Helmet>
        <title>Kern Zhou - Work Experience</title>
        <meta property="og:title" content="Kern Zhou - Work Experience" />
        <meta name="description" content="Kern Zhou's professional work experience and technical development journey" />
      </Helmet>

      <Hero17
        heading1={wrap(data.hero['heading1'])}
        content1={wrap(data.hero['content1'])}
        action1={wrap(data.hero['action1'])}
        action2={wrap(data.hero['action2'])}
      />

      <Features24
        feature1Title={wrap(data.features24['feature1Title'])}
        feature2Title={wrap(data.features24['feature2Title'])}
        feature3Title={wrap(data.features24['feature3Title'])}
        feature1Description={wrap(data.features24['feature1Description'])}
        feature2Description={wrap(data.features24['feature2Description'])}
        feature3Description={wrap(data.features24['feature3Description'])}
      />

      <Features25
        feature1Title={wrap(data.features25['feature1Title'])}
        feature2Title={wrap(data.features25['feature2Title'])}
        feature3Title={wrap(data.features25['feature3Title'])}
        feature1Description={wrap(data.features25['feature1Description'])}
        feature2Description={wrap(data.features25['feature2Description'])}
        feature3Description={wrap(data.features25['feature3Description'])}
      />

      <Steps2
        step1Title={wrap(data.steps2['step1Title'])}
        step2Title={wrap(data.steps2['step2Title'])}
        step3Title={wrap(data.steps2['step3Title'])}
        step4Title={wrap(data.steps2['step4Title'])}
        step1Description={wrap(data.steps2['step1Description'])}
        step2Description={wrap(data.steps2['step2Description'])}
        step3Description={wrap(data.steps2['step3Description'])}
        step4Description={wrap(data.steps2['step4Description'])}
      />

      <Contact10
        heading1={wrap(data.contact10['heading1'])}
        content1={wrap(data.contact10['content1'])}
        email1={wrap(data.contact10['email1'])}
      />
    </div>
  );
};

export default ZsWorkExperience;
