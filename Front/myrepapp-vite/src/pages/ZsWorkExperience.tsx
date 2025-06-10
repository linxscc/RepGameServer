import React, { Fragment } from 'react';
import { Helmet } from 'react-helmet';
import './ZsWorkExperience.css';

// 导入工作经验组件
import Hero17 from '@/assets/work-experience/hero17';
import Features24 from '@/assets/work-experience/features24';
import Features25 from '@/assets/work-experience/features25';
import Steps2 from '@/assets/work-experience/steps2';
import Contact10 from '@/assets/work-experience/contact10';

// 导入样式
import '@/assets/work-experience/global-style.css';
import '@/assets/work-experience/home.css';

const ZsWorkExperience: React.FC = () => {
  return (
    <div className="home-container">
      <Helmet>
        <title>Kern Zhou - Work Experience</title>
        <meta property="og:title" content="Kern Zhou - Work Experience" />
        <meta name="description" content="Kern Zhou's professional work experience and technical development journey, including game development, enterprise systems, and database optimization experience" />
      </Helmet>
      
      <Hero17
        heading1={
          <Fragment>
            <span>Kern Zhou - Software Engineer</span>
          </Fragment>
        }
        content1={
          <Fragment>
            <span>
              Years of software development experience, from game development to enterprise systems, proficient in multiple technology stacks and development processes
            </span>
          </Fragment>
        }
        action1={
          <Fragment>
            <span>View Project Experience</span>
          </Fragment>
        }
        action2={
          <Fragment>
            <span>Contact Me</span>
          </Fragment>
        }
      />
      
      <Features24
        feature1Title={
          <Fragment>
            <span>Backend Development Expertise</span>
          </Fragment>
        }
        feature2Title={
          <Fragment>
            <span>Frontend Technology Stack</span>
          </Fragment>
        }
        feature3Title={
          <Fragment>
            <span>Database Optimization</span>
          </Fragment>
        }
        feature1Description={
          <Fragment>
            <span>
              Proficient in Python Django, C#, VB.Net and other backend technologies, with extensive enterprise system development experience
            </span>
          </Fragment>
        }
        feature2Description={
          <Fragment>
            <span>
              Skilled in Unity, Vuetify, React and other frontend technologies, with complete full-stack development capabilities
            </span>
          </Fragment>
        }
        feature3Description={
          <Fragment>
            <span>
              Professional SQL optimization skills, processed tens of millions of data records, ensuring high-performance system operation
            </span>
          </Fragment>
        }
      />
      
      <Features25
        feature1Title={
          <Fragment>
            <span>Enterprise System Development</span>
          </Fragment>
        }
        feature2Title={
          <Fragment>
            <span>Game Development Experience</span>
          </Fragment>
        }
        feature3Title={
          <Fragment>
            <span>Multilingual Environment</span>
          </Fragment>
        }
        feature1Description={
          <Fragment>
            <span>
              Experience in developing large-scale enterprise applications including ERP systems, banking systems, and government invoice systems
            </span>
          </Fragment>
        }
        feature2Description={
          <Fragment>
            <span>
              Unity 3D game development, including complete implementation of character animation, skill systems, and weapon/prop systems
            </span>
          </Fragment>
        }
        feature3Description={
          <Fragment>
            <span>
              Native Chinese speaker, Japanese N1 level, English B1 level, capable of international project development
            </span>
          </Fragment>
        }
      />
      
      <Steps2
        step1Title={
          <Fragment>
            <span>Game Development Engineer (2018-2020)</span>
          </Fragment>
        }
        step2Title={
          <Fragment>
            <span>Database Engineer (2020-2022)</span>
          </Fragment>
        }
        step3Title={
          <Fragment>
            <span>Desktop Development Engineer (2022-2023)</span>
          </Fragment>
        }
        step4Title={
          <Fragment>
            <span>System Engineer (2024-Present)</span>
          </Fragment>
        }
        step1Description={
          <Fragment>
            <span>
              XISHANJU - MMOARPG mobile game development, using Unity and C# to develop core features including character animation and skill systems
            </span>
          </Fragment>
        }
        step2Description={
          <Fragment>
            <span>
              HITACHI - Banking system backend development, using C# and SQL to process tens of millions of data records, implementing APIs and automated testing
            </span>
          </Fragment>
        }
        step3Description={
          <Fragment>
            <span>
              HITACHI - Government invoice system, using VB.Net to develop Windows desktop applications, adapting to Japanese tax policies
            </span>
          </Fragment>
        }
        step4Description={
          <Fragment>
            <span>
              HITACHI/SUPER YACHT GROUP - ERP systems and customer management systems, using Python Django and Vuetify
            </span>
          </Fragment>
        }
      />
      
      <Contact10
        content1={
          <Fragment>
            <span>
              If you are interested in my work experience or have collaboration opportunities, please feel free to contact me through the following methods
            </span>
          </Fragment>
        }
        heading1={
          <Fragment>
            <span>Contact Kern Zhou</span>
          </Fragment>
        }
        email1={
          <Fragment>
            <span>kern.zhou1995@gmail.com</span>
          </Fragment>
        }
      />
    </div>
  );
};

export default ZsWorkExperience;
